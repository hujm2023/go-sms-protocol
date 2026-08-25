package main

import (
	"context"
	cryptorand "crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/hujm2023/hlog"

	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/cmpp/cmpp20"
	"github.com/hujm2023/go-sms-protocol/nioserver"
)

type receiptJob struct {
	conn          nioserver.ISMSConn[connectionSession]
	messageID     uint64
	deliveryID    uint64
	destination   string
	destinationID string
	serviceID     string
	submittedAt   time.Time
	smscSequence  uint32
}

// receiptDispatcher 用固定数量的 worker 和有界队列模拟异步回执。
// 这样即使客户端持续高速提交，也不会为每条短信创建 goroutine 或无限占用内存。
type receiptDispatcher struct {
	ctx          context.Context
	logger       hlog.FullLogger
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
	logger hlog.FullLogger,
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
		d.logger.Errorf("build delivery receipt msg_id=%s: %v", cmpp.MsgID2String(job.messageID), err)
		return
	}
	data, err := delivery.IEncode()
	if err != nil {
		d.logger.Errorf("encode delivery receipt msg_id=%s: %v", cmpp.MsgID2String(job.messageID), err)
		return
	}

	writeCtx, cancel := context.WithTimeout(d.ctx, d.writeTimeout)
	err = job.conn.Write(writeCtx, data)
	cancel()
	if err != nil {
		if d.ctx.Err() == nil && !errors.Is(err, nioserver.ErrConnectionClosed) {
			d.logger.Errorf("push delivery receipt msg_id=%s remote=%s: %v", cmpp.MsgID2String(job.messageID), job.conn.RemoteAddr(), err)
		}
		return
	}
	d.logger.Infof(
		"delivery receipt pushed remote=%s submit_msg_id=%s delivery_msg_id=%s sequence=%d status=%s",
		job.conn.RemoteAddr(),
		cmpp.MsgID2String(job.messageID),
		cmpp.MsgID2String(job.deliveryID),
		delivery.GetSequenceID(),
		cmpp20.DELIVERED,
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

func buildDeliveryReceipt(job receiptJob, doneAt time.Time, sequenceID uint32) (*cmpp20.PduDeliver, error) {
	// 状态报告里有两个不同的 MsgID：
	// 1. 外层 PduDeliver.MsgID 是本次推送的新 ID，客户端应在 DELIVER_RESP 中回显；
	// 2. 回执正文 MsgID 是原 SUBMIT_RESP 的 ID，用来关联最初提交。
	content, err := (&cmpp.SubPduDeliveryContent{
		MsgID:          job.messageID,
		Stat:           cmpp20.DELIVERED,
		SubmitTime:     job.submittedAt.Format("0601021504"),
		DoneTime:       doneAt.Format("0601021504"),
		DestTerminalID: job.destination,
		SMSCSequence:   job.smscSequence,
	}).IEncode()
	if err != nil {
		return nil, fmt.Errorf("encode delivery content: %w", err)
	}
	if len(content) > 255 {
		return nil, fmt.Errorf("delivery content length %d exceeds one byte", len(content))
	}

	return &cmpp20.PduDeliver{
		Header:            cmpp.NewHeader(0, cmpp.CommandDeliver, sequenceID),
		MsgID:             job.deliveryID,
		DestID:            job.destinationID,
		ServiceID:         job.serviceID,
		SrcTerminalID:     job.destination,
		RegisteredDeliver: 1,
		MsgLength:         uint8(len(content)),
		MsgContent:        content,
	}, nil
}

// heartbeatManager 用一个定时器扫描全部已鉴权连接。
// 相比“每连接一个 ticker”，这种方式更容易统一停止，也能避免连接数增长时创建大量定时器。
type heartbeatManager struct {
	ctx          context.Context
	logger       hlog.FullLogger
	interval     time.Duration
	maxMissed    int
	writeTimeout time.Duration
	mu           sync.RWMutex
	connections  map[nioserver.ISMSConn[connectionSession]]struct{}
	done         chan struct{}
}

func newHeartbeatManager(
	ctx context.Context,
	logger hlog.FullLogger,
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

	request := &cmpp20.PduActiveTest{
		Header: cmpp.NewHeader(0, cmpp.CommandActiveTest, conn.NextSequenceID()),
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
