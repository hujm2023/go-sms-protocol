package main

import (
	"context"
	"errors"
	"sync"
	"time"

	smslog "github.com/hujm2023/go-sms-protocol/logger"
	"github.com/hujm2023/go-sms-protocol/nioserver"
	"github.com/hujm2023/go-sms-protocol/smpp"
	"github.com/hujm2023/go-sms-protocol/smpp/smpp34"
)

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

	request := &smpp34.EnquireLink{
		Header: smpp.Header{ID: smpp.ENQUIRE_LINK, Sequence: nextSequence(conn)},
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
