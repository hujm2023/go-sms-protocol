package main

import (
	"context"
	cryptorand "crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	smslog "github.com/hujm2023/go-sms-protocol/logger"
	"github.com/hujm2023/go-sms-protocol/nioserver"
	"github.com/hujm2023/go-sms-protocol/smgp"
	"github.com/hujm2023/go-sms-protocol/smgp/smgp30"
)

type receiptJob struct {
	conn          nioserver.ISMSConn[connectionSession]
	messageID     string
	deliveryID    string
	destination   string
	destinationID string
	submittedAt   time.Time
}

// receiptDispatcher 用固定数量的 worker 和有界队列模拟异步回执。
// 这样即使客户端持续高速提交，也不会为每条短信创建 goroutine 或无限占用内存。
type receiptDispatcher struct {
	ctx          context.Context
	logger       smslog.FullLogger
	minDelay     time.Duration
	maxDelay     time.Duration
	writeTimeout time.Duration
	jobs         chan receiptJob
	enqueueMu    sync.Mutex
	done         chan struct{}
	now          func() time.Time
	delay        func() time.Duration
}

func newReceiptDispatcher(
	ctx context.Context,
	logger smslog.FullLogger,
	workerCount int,
	queueSize int,
	minDelay time.Duration,
	maxDelay time.Duration,
	writeTimeout time.Duration,
) *receiptDispatcher {
	dispatcher := &receiptDispatcher{
		ctx:          ctx,
		logger:       logger,
		minDelay:     minDelay,
		maxDelay:     maxDelay,
		writeTimeout: writeTimeout,
		jobs:         make(chan receiptJob, queueSize),
		done:         make(chan struct{}),
		now:          time.Now,
	}
	dispatcher.delay = dispatcher.randomDelay

	// worker 的生命周期绑定到服务级 ctx，而不是某一次请求的 ctx。
	// handler 返回后，请求 ctx 可能被取消，但已经受理的回执仍需要按约定异步推送。
	var workers sync.WaitGroup
	workers.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go func() {
			defer workers.Done()
			dispatcher.runWorker()
		}()
	}
	go func() {
		workers.Wait()
		close(dispatcher.done)
	}()
	return dispatcher
}

func (d *receiptDispatcher) enqueue(jobs []receiptJob) bool {
	if len(jobs) == 0 {
		return true
	}

	// 一个 SUBMIT 可以包含多个接收号码。这里串行化“检查容量 + 整批入队”，
	// 保证要么所有号码都安排回执，要么以流控错误拒绝整个 SUBMIT，避免部分成功。
	d.enqueueMu.Lock()
	defer d.enqueueMu.Unlock()
	if d.ctx.Err() != nil || len(jobs) > cap(d.jobs)-len(d.jobs) {
		return false
	}
	for _, job := range jobs {
		d.jobs <- job
	}
	return true
}

func (d *receiptDispatcher) runWorker() {
	for {
		if d.ctx.Err() != nil {
			return
		}
		select {
		case <-d.ctx.Done():
			return
		case job := <-d.jobs:
			d.deliver(job)
		}
	}
}

func (d *receiptDispatcher) deliver(job receiptJob) {
	// Timer 可被服务关闭信号打断，避免优雅退出等待完整的模拟延迟。
	timer := time.NewTimer(d.delay())
	defer timer.Stop()
	select {
	case <-d.ctx.Done():
		return
	case <-timer.C:
	}

	delivery, err := buildDeliveryReceipt(job, d.now(), job.conn.NextSequenceID())
	if err != nil {
		d.logger.Errorf("build delivery receipt msg_id=%s: %v", job.messageID, err)
		return
	}
	data, err := delivery.IEncode()
	if err != nil {
		d.logger.Errorf("encode delivery receipt msg_id=%s: %v", job.messageID, err)
		return
	}

	writeCtx, cancel := context.WithTimeout(d.ctx, d.writeTimeout)
	err = job.conn.Write(writeCtx, data)
	cancel()
	if err != nil {
		if d.ctx.Err() == nil && !errors.Is(err, nioserver.ErrConnectionClosed) {
			d.logger.Errorf("push delivery receipt msg_id=%s remote=%s: %v", job.messageID, job.conn.RemoteAddr(), err)
		}
		return
	}
	d.logger.Infof(
		"delivery receipt pushed remote=%s submit_msg_id=%s delivery_msg_id=%s sequence=%d status=%s",
		job.conn.RemoteAddr(),
		job.messageID,
		job.deliveryID,
		delivery.GetSequenceID(),
		"DELIVRD",
	)
}

func (d *receiptDispatcher) randomDelay() time.Duration {
	span := d.maxDelay - d.minDelay
	if span <= 0 {
		return d.minDelay
	}
	// 使用 crypto/rand 不需要维护共享伪随机源，也不会让并发 worker 因错误的随机源用法产生数据竞争。
	n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(span)+1))
	if err != nil {
		d.logger.Warnf("generate receipt delay: %v", err)
		return d.minDelay
	}
	return d.minDelay + time.Duration(n.Int64())
}

func (d *receiptDispatcher) wait(ctx context.Context) error {
	select {
	case <-d.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func buildDeliveryReceipt(job receiptJob, doneAt time.Time, sequenceID uint32) (*smgp30.Deliver, error) {
	// 正文 id 是原 SUBMIT_RESP 的 10 字节二进制 MsgID；外层 MsgID 标识本次推送。
	content, err := (&smgp30.DeliveryReceipt{
		ID:       job.messageID,
		Sub:      "001",
		Dlvrd:    "001",
		SubDate:  job.submittedAt.Format("0601021504"),
		DoneDate: doneAt.Format("0601021504"),
		Stat:     "DELIVRD",
		Err:      "000",
		Text:     "000" + strings.Repeat(" ", 17),
	}).IEncode()
	if err != nil {
		return nil, fmt.Errorf("encode delivery content: %w", err)
	}
	return &smgp30.Deliver{
		Header:     smgp.NewHeader(0, smgp.CommandDeliver, sequenceID),
		MsgID:      job.deliveryID,
		IsReport:   smgp.IS_REPORT,
		MsgFormat:  smgp.ASCII,
		RecvTime:   doneAt.Format("20060102150405"),
		SrcTermID:  job.destination,
		DestTermID: job.destinationID,
		MsgLength:  uint8(len(content)),
		MsgContent: content,
	}, nil
}

// heartbeatManager 用一个定时器扫描全部已鉴权连接。
// 相比“每连接一个 ticker”，这种方式更容易统一停止，也能避免连接数增长时创建大量定时器。
type heartbeatManager struct {
	ctx          context.Context
	logger       smslog.FullLogger
	interval     time.Duration
	maxMissed    int
	writeTimeout time.Duration
	mu           sync.RWMutex
	connections  map[nioserver.ISMSConn[connectionSession]]struct{}
	done         chan struct{}
}

func newHeartbeatManager(
	ctx context.Context,
	logger smslog.FullLogger,
	interval time.Duration,
	maxMissed int,
	writeTimeout time.Duration,
) *heartbeatManager {
	manager := &heartbeatManager{
		ctx:          ctx,
		logger:       logger,
		interval:     interval,
		maxMissed:    maxMissed,
		writeTimeout: writeTimeout,
		connections:  make(map[nioserver.ISMSConn[connectionSession]]struct{}),
		done:         make(chan struct{}),
	}
	go manager.run()
	return manager
}

func (m *heartbeatManager) register(conn nioserver.ISMSConn[connectionSession]) {
	if m.ctx.Err() != nil {
		return
	}
	m.mu.Lock()
	m.connections[conn] = struct{}{}
	m.mu.Unlock()
}

func (m *heartbeatManager) unregister(conn nioserver.ISMSConn[connectionSession]) {
	m.mu.Lock()
	delete(m.connections, conn)
	m.mu.Unlock()
}

func (m *heartbeatManager) run() {
	defer close(m.done)
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.checkConnections()
		}
	}
}

func (m *heartbeatManager) checkConnections() {
	// 网络写入不能持有连接表的读锁，否则一个慢连接会阻塞连接关闭时的 unregister。
	// 因此先复制快照，再逐个探测。
	m.mu.RLock()
	connections := make([]nioserver.ISMSConn[connectionSession], 0, len(m.connections))
	for conn := range m.connections {
		connections = append(connections, conn)
	}
	m.mu.RUnlock()

	for _, conn := range connections {
		if m.ctx.Err() != nil {
			return
		}
		m.checkConnection(conn)
	}
}

func (m *heartbeatManager) checkConnection(conn nioserver.ISMSConn[connectionSession]) {
	if conn.NoActiveTestCount() >= m.maxMissed {
		m.logger.Warnf(
			"heartbeat timeout remote=%s unanswered=%d",
			conn.RemoteAddr(),
			conn.NoActiveTestCount(),
		)
		if err := conn.Close(); err != nil && !errors.Is(err, nioserver.ErrConnectionClosed) {
			m.logger.Errorf("close heartbeat-timeout connection remote=%s: %v", conn.RemoteAddr(), err)
		}
		m.unregister(conn)
		return
	}

	request := &smgp30.ActiveTest{
		Header: smgp.NewHeader(0, smgp.CommandActiveTest, conn.NextSequenceID()),
	}
	data, err := request.IEncode()
	if err != nil {
		m.logger.Errorf("encode heartbeat remote=%s: %v", conn.RemoteAddr(), err)
		return
	}
	writeCtx, cancel := context.WithTimeout(m.ctx, m.writeTimeout)
	err = conn.Write(writeCtx, data)
	cancel()
	if err != nil {
		if m.ctx.Err() == nil && !errors.Is(err, nioserver.ErrConnectionClosed) {
			m.logger.Errorf("write heartbeat remote=%s: %v", conn.RemoteAddr(), err)
		}
		if closeErr := conn.Close(); closeErr != nil && !errors.Is(closeErr, nioserver.ErrConnectionClosed) {
			m.logger.Errorf("close heartbeat-write connection remote=%s: %v", conn.RemoteAddr(), closeErr)
		}
		m.unregister(conn)
		return
	}
	// 只有探活请求真正写出后才增加未应答次数，传输失败会直接关闭连接，不能重复计数。
	conn.OnSendActiveTest()
}

func (m *heartbeatManager) wait(ctx context.Context) error {
	select {
	case <-m.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
