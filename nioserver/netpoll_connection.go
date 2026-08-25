package nioserver

import (
	"context"
	"errors"
	"io"
	"math"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/cloudwego/netpoll"
)

// IActiveTest defines the active-test state maintained for a connection.
type IActiveTest interface {
	// OnSendActiveTest records a sent active-test request.
	OnSendActiveTest()
	// OnReceiveActiveTest resets the counter after an active-test response.
	OnReceiveActiveTest()
	// NoActiveTestCount returns the number of consecutive unanswered requests.
	NoActiveTestCount() int
}

// ISMSConn defines an SMS protocol connection with generic business data.
//
// GetBizData and SetBizData synchronize replacement of the value itself. They
// do not make maps, slices, pointers, or other mutable objects stored in T safe
// for concurrent use.
type ISMSConn[T any] interface {
	IActiveTest
	io.Closer

	// Write serializes and flushes data to the peer. The caller may reuse data
	// after Write returns.
	Write(ctx context.Context, data []byte) error
	// WriteAndClose flushes data before closing the connection.
	WriteAndClose(ctx context.Context, data []byte) error

	// AsyncWrite is retained for source compatibility.
	// Deprecated: use Write and handle the returned error.
	AsyncWrite(ctx context.Context, data []byte)

	// RemoteAddr returns the cached remote network address.
	RemoteAddr() string
	// NextSequenceID returns the next non-zero outbound sequence ID.
	NextSequenceID() uint32

	// GetBizData returns the business data associated with this connection.
	GetBizData() T
	// SetBizData replaces the business data associated with this connection.
	SetBizData(data T)
}

// connKey is private to prevent context-key collisions.
type connKey struct{}

var ctxKey connKey

// muxConn is the connection-scoped session used by BaseServer. Despite the
// historical name, writes are no longer owned by netpoll/mux.ShardQueue.
type muxConn[T any] struct {
	conn        netpoll.Connection
	server      *BaseServer[T]
	writer      *serialWriter
	handlerGate *admissionGate

	ctx    context.Context
	cancel context.CancelCauseFunc

	closeOnce          sync.Once
	transportCloseOnce sync.Once
	cleanupDone        chan struct{}

	sequenceIDGen atomic.Uint32
	noActiveTest  atomic.Uint32
	remoteAddr    string

	bizMu   sync.RWMutex
	bizData T
}

// newSvrMuxConn creates a server-side connection session. The caller must
// register the transport close bridge before exposing the returned session.
func newSvrMuxConn[T any](conn netpoll.Connection, server *BaseServer[T]) *muxConn[T] {
	ctx, cancel := context.WithCancelCause(context.Background())
	mc := &muxConn[T]{
		conn:        conn,
		server:      server,
		ctx:         ctx,
		cancel:      cancel,
		cleanupDone: make(chan struct{}),
		handlerGate: newAdmissionGate(server.maxHandlersPerConn),
	}
	if addr := conn.RemoteAddr(); addr != nil {
		mc.remoteAddr = addr.String()
	}
	mc.writer = newSerialWriter(server.maxPendingWrites, mc.flush)
	return mc
}

func (m *muxConn[T]) flush(data []byte) error {
	if !m.conn.IsActive() {
		return ErrConnectionClosed
	}
	w := m.conn.Writer()
	buf, err := w.Malloc(len(data))
	if err != nil {
		return err
	}
	copy(buf, data)
	if err := w.Flush(); err != nil {
		return err
	}
	return nil
}

// Write copies, serializes, and flushes data through the connection writer.
func (m *muxConn[T]) Write(ctx context.Context, data []byte) error {
	return m.writer.Write(ctx, data)
}

// WriteAndClose flushes data before beginning the common close path.
func (m *muxConn[T]) WriteAndClose(ctx context.Context, data []byte) error {
	writeErr := m.Write(ctx, data)
	closeErr := m.Close()
	return errors.Join(writeErr, closeErr)
}

// AsyncWrite preserves the old API while routing through the safe writer.
// Deprecated: use Write and handle the returned error.
func (m *muxConn[T]) AsyncWrite(ctx context.Context, data []byte) {
	if err := m.Write(ctx, data); err != nil {
		m.server.logger.CtxErrorf(ctx, "write response error: %v", err)
	}
}

func (m *muxConn[T]) RemoteAddr() string {
	return m.remoteAddr
}

// NextSequenceID atomically returns the next non-zero sequence ID.
func (m *muxConn[T]) NextSequenceID() uint32 {
	n := m.sequenceIDGen.Add(1)
	if n == 0 {
		n = m.sequenceIDGen.Add(1)
	}
	return n
}

func (m *muxConn[T]) GetBizData() T {
	m.bizMu.RLock()
	defer m.bizMu.RUnlock()
	return m.bizData
}

func (m *muxConn[T]) SetBizData(data T) {
	m.bizMu.Lock()
	m.bizData = data
	m.bizMu.Unlock()
}

// OnSendActiveTest increments the missed-response count and saturates at
// math.MaxUint32 so overflow never makes an unhealthy connection look healthy.
func (m *muxConn[T]) OnSendActiveTest() {
	limit := maxActiveTestCount()
	for {
		current := m.noActiveTest.Load()
		if current == limit {
			return
		}
		if m.noActiveTest.CompareAndSwap(current, current+1) {
			return
		}
	}
}

func maxActiveTestCount() uint32 {
	if strconv.IntSize == 32 {
		return math.MaxInt32
	}
	return math.MaxUint32
}

func (m *muxConn[T]) NoActiveTestCount() int {
	return int(m.noActiveTest.Load())
}

func (m *muxConn[T]) OnReceiveActiveTest() {
	m.noActiveTest.Store(0)
}

// beginClose is phase A. It is deliberately constant-time and safe to invoke
// from netpoll's CloseCallback chain.
func (m *muxConn[T]) beginClose(cause error) {
	if cause == nil {
		cause = ErrConnectionClosed
	}
	m.closeOnce.Do(func() {
		m.handlerGate.Close()
		m.writer.BeginStop()
		m.cancel(cause)
	})
}

// transportClosed schedules phase B exactly once. It never runs user code on
// the netpoll callback goroutine.
func (m *muxConn[T]) transportClosed(cause error) {
	m.beginClose(cause)
	m.transportCloseOnce.Do(func() {
		m.server.scheduleCleanup(m)
	})
}

func (m *muxConn[T]) closeWithCause(cause error) error {
	m.beginClose(cause)
	err := m.conn.Close()
	// netpoll may defer CloseCallback while OnRequest holds its processing
	// lock. Phase B must not depend on that callback running promptly.
	m.transportClosed(cause)
	return err
}

func (m *muxConn[T]) Close() error {
	return m.closeWithCause(ErrConnectionClosed)
}

func (m *muxConn[T]) connectionContext() context.Context {
	return m.ctx
}

// fillCtx adds the connection session to ctx.
func fillCtx[T any](ctx context.Context, conn ISMSConn[T]) context.Context {
	return context.WithValue(ctx, ctxKey, conn)
}

// GetCtxConn extracts the connection session from ctx.
func GetCtxConn[T any](ctx context.Context) (ISMSConn[T], bool) {
	if ctx == nil {
		return nil, false
	}
	conn, ok := ctx.Value(ctxKey).(ISMSConn[T])
	return conn, ok
}
