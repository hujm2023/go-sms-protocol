package nioserver

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/cloudwego/netpoll"

	protocol "github.com/hujm2023/go-sms-protocol"
)

func TestMuxConnBusinessDataSupportsInterfaceNilAndDynamicTypes(t *testing.T) {
	server := newConnectionTestServer[any](t)
	conn := newFakeConnection()
	mc := newSvrMuxConn[any](conn, server)

	if got := mc.GetBizData(); got != nil {
		t.Fatalf("initial business data = %#v, want nil", got)
	}
	mc.SetBizData("value")
	if got := mc.GetBizData(); got != "value" {
		t.Fatalf("string business data = %#v, want value", got)
	}
	mc.SetBizData(42)
	if got := mc.GetBizData(); got != 42 {
		t.Fatalf("integer business data = %#v, want 42", got)
	}
	mc.SetBizData(nil)
	if got := mc.GetBizData(); got != nil {
		t.Fatalf("nil replacement = %#v, want nil", got)
	}
}

func TestMuxConnBusinessDataConcurrentReplacement(t *testing.T) {
	server := newConnectionTestServer[*int](t)
	mc := newSvrMuxConn[*int](newFakeConnection(), server)

	const goroutines = 16
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		value := i
		go func() {
			defer wg.Done()
			mc.SetBizData(&value)
			_ = mc.GetBizData()
		}()
	}
	wg.Wait()
	if mc.GetBizData() == nil {
		t.Fatal("business data unexpectedly nil after concurrent replacement")
	}
}

func TestMuxConnActiveTestSaturatesAndResets(t *testing.T) {
	server := newConnectionTestServer[any](t)
	mc := newSvrMuxConn[any](newFakeConnection(), server)

	mc.OnSendActiveTest()
	mc.OnSendActiveTest()
	if got := mc.NoActiveTestCount(); got != 2 {
		t.Fatalf("NoActiveTestCount() = %d, want 2", got)
	}
	mc.OnReceiveActiveTest()
	if got := mc.NoActiveTestCount(); got != 0 {
		t.Fatalf("count after response = %d, want 0", got)
	}
	limit := maxActiveTestCount()
	mc.noActiveTest.Store(limit - 1)
	mc.OnSendActiveTest()
	mc.OnSendActiveTest()
	if got := mc.noActiveTest.Load(); got != limit {
		t.Fatalf("saturated count = %d, want %d", got, limit)
	}
}

func TestMuxConnCloseRunsPhaseAImmediatelyAndCleanupOnce(t *testing.T) {
	var callbackCount atomic.Int32
	callbackObserved := make(chan struct {
		connectionInactive bool
		cleanupContextErr  error
	}, 1)
	server := newConnectionTestServer[any](t, WithOnCloseFunc[any](func(ctx context.Context, conn netpoll.Connection) {
		callbackCount.Add(1)
		callbackObserved <- struct {
			connectionInactive bool
			cleanupContextErr  error
		}{
			connectionInactive: !conn.IsActive(),
			cleanupContextErr:  ctx.Err(),
		}
	}))
	setServerAccepting(server)

	conn := newFakeConnection()
	conn.delayCallbacks = true
	ctx := server.OnOpenConn(conn)
	mc, ok := sessionFromContext[any](ctx)
	if !ok {
		t.Fatal("OnOpenConn context is missing muxConn")
	}

	const closers = 16
	var wg sync.WaitGroup
	wg.Add(closers)
	for i := 0; i < closers; i++ {
		go func() {
			defer wg.Done()
			_ = mc.Close()
		}()
	}
	wg.Wait()

	if !errors.Is(context.Cause(mc.connectionContext()), ErrConnectionClosed) {
		t.Fatalf("connection context cause = %v, want ErrConnectionClosed", context.Cause(mc.connectionContext()))
	}
	if err := mc.Write(context.Background(), []byte("late")); !errors.Is(err, ErrConnectionClosed) {
		t.Fatalf("Write after Close = %v, want ErrConnectionClosed", err)
	}
	if err := mc.handlerGate.Acquire(context.Background()); !errors.Is(err, ErrAdmissionClosed) {
		t.Fatalf("handler admission after Close = %v, want ErrAdmissionClosed", err)
	}

	observation := receiveCloseObservation(t, callbackObserved)
	if !observation.connectionInactive {
		t.Fatal("OnCloseFunc observed an active transport")
	}
	if observation.cleanupContextErr != nil {
		t.Fatalf("cleanup context was already canceled: %v", observation.cleanupContextErr)
	}

	// The real netpoll callback may be delayed until OnRequest releases its
	// processing lock. Firing the stored callback later must not duplicate
	// phase B.
	conn.fireCloseCallbacks()
	waitForChannel(t, mc.cleanupDone, "connection cleanup")
	if got := callbackCount.Load(); got != 1 {
		t.Fatalf("OnCloseFunc calls = %d, want 1", got)
	}
	server.cleanupTracker.Close()
	if err := server.cleanupTracker.Wait(context.Background()); err != nil {
		t.Fatalf("cleanup tracker Wait: %v", err)
	}
}

func TestMuxConnWriteAndCloseFlushesCopiedDataBeforeClose(t *testing.T) {
	server := newConnectionTestServer[any](t)
	setServerAccepting(server)
	conn := newFakeConnection()
	ctx := server.OnOpenConn(conn)
	mc, ok := sessionFromContext[any](ctx)
	if !ok {
		t.Fatal("OnOpenConn context is missing muxConn")
	}

	data := []byte("response")
	if err := mc.WriteAndClose(context.Background(), data); err != nil {
		t.Fatalf("WriteAndClose() error = %v", err)
	}
	copy(data, "modified")

	if got := string(conn.writtenBytes()); got != "response" {
		t.Fatalf("wire data = %q, want response", got)
	}
	if got := conn.eventSequence(); len(got) < 2 || got[0] != "flush" || got[1] != "close" {
		t.Fatalf("event sequence = %v, want flush before close", got)
	}
	waitForChannel(t, mc.cleanupDone, "connection cleanup")
	server.cleanupTracker.Close()
}

func TestMuxConnCleanupIsolatesUserCallbackPanic(t *testing.T) {
	callbackEntered := make(chan struct{})
	server := newConnectionTestServer[any](t, WithOnCloseFunc[any](func(context.Context, netpoll.Connection) {
		close(callbackEntered)
		panic("callback panic")
	}))
	setServerAccepting(server)
	ctx := server.OnOpenConn(newFakeConnection())
	mc, ok := sessionFromContext[any](ctx)
	if !ok {
		t.Fatal("OnOpenConn context is missing muxConn")
	}

	if err := mc.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	waitForChannel(t, callbackEntered, "close callback entry")
	waitForChannel(t, mc.cleanupDone, "connection cleanup after callback panic")
	server.cleanupTracker.Close()
	if err := server.cleanupTracker.Wait(context.Background()); err != nil {
		t.Fatalf("cleanup tracker Wait: %v", err)
	}
}

func TestMuxConnBlockingUserCallbackDoesNotBlockCloseBridge(t *testing.T) {
	callbackEntered := make(chan struct{})
	releaseCallback := make(chan struct{})
	server := newConnectionTestServer[any](t, WithOnCloseFunc[any](func(context.Context, netpoll.Connection) {
		close(callbackEntered)
		<-releaseCallback
	}))
	setServerAccepting(server)
	conn := newFakeConnection()
	conn.delayCallbacks = true
	ctx := server.OnOpenConn(conn)
	mc, ok := sessionFromContext[any](ctx)
	if !ok {
		t.Fatal("OnOpenConn context is missing muxConn")
	}

	closeResult := make(chan error, 1)
	go func() { closeResult <- mc.Close() }()
	if err := receiveError(t, closeResult, "Close result"); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	waitForChannel(t, callbackEntered, "blocking close callback entry")

	bridgeReturned := make(chan struct{})
	go func() {
		conn.fireCloseCallbacks()
		close(bridgeReturned)
	}()
	waitForChannel(t, bridgeReturned, "delayed netpoll close bridge")
	select {
	case <-mc.cleanupDone:
		t.Fatal("cleanup completed while user callback was blocked")
	default:
	}

	close(releaseCallback)
	waitForChannel(t, mc.cleanupDone, "cleanup after callback release")
	server.cleanupTracker.Close()
	if err := server.cleanupTracker.Wait(context.Background()); err != nil {
		t.Fatalf("cleanup tracker Wait: %v", err)
	}
}

func newConnectionTestServer[T any](t *testing.T, opts ...ServerOption[T]) *BaseServer[T] {
	t.Helper()
	baseOptions := []ServerOption[T]{
		WithUnpackFunc[T](func(context.Context, netpoll.Reader) (protocol.PDU, error) {
			return nil, nil
		}),
		WithHandleFunc[T](func(context.Context, protocol.PDU) ([]byte, error) {
			return nil, nil
		}),
	}
	baseOptions = append(baseOptions, opts...)
	server, err := NewBaseServer[T]("tcp", "127.0.0.1:0", baseOptions...)
	if err != nil {
		t.Fatalf("NewBaseServer() error = %v", err)
	}
	return server
}

func setServerAccepting[T any](server *BaseServer[T]) {
	server.sessionsMu.Lock()
	server.accepting = true
	server.sessionsMu.Unlock()
}

func receiveCloseObservation[T any](t *testing.T, observations <-chan T) T {
	t.Helper()
	select {
	case observation := <-observations:
		return observation
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for close callback observation")
		var zero T
		return zero
	}
}

func waitForChannel(t *testing.T, done <-chan struct{}, description string) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for %s", description)
	}
}

type fakeConnection struct {
	mu             sync.Mutex
	active         bool
	delayCallbacks bool
	callbacks      []netpoll.CloseCallback
	callbacksFired bool
	closeCount     int
	events         []string
	wire           bytes.Buffer
	reader         netpoll.Reader
	writer         *fakeNetpollWriter
}

func newFakeConnection() *fakeConnection {
	conn := &fakeConnection{
		active: true,
		reader: netpoll.NewReader(bytes.NewReader(nil)),
	}
	conn.writer = &fakeNetpollWriter{conn: conn}
	return conn
}

func (c *fakeConnection) Read([]byte) (int, error) { return 0, io.EOF }

func (c *fakeConnection) Write(data []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.wire.Write(data)
}

func (c *fakeConnection) Close() error {
	c.mu.Lock()
	if !c.active {
		c.mu.Unlock()
		return nil
	}
	c.active = false
	c.closeCount++
	c.events = append(c.events, "close")
	delayed := c.delayCallbacks
	c.mu.Unlock()
	if !delayed {
		c.fireCloseCallbacks()
	}
	return nil
}

func (c *fakeConnection) LocalAddr() net.Addr  { return fakeAddr("local") }
func (c *fakeConnection) RemoteAddr() net.Addr { return fakeAddr("remote") }

func (c *fakeConnection) SetDeadline(time.Time) error      { return nil }
func (c *fakeConnection) SetReadDeadline(time.Time) error  { return nil }
func (c *fakeConnection) SetWriteDeadline(time.Time) error { return nil }
func (c *fakeConnection) Reader() netpoll.Reader           { return c.reader }
func (c *fakeConnection) Writer() netpoll.Writer           { return c.writer }

func (c *fakeConnection) IsActive() bool {
	c.mu.Lock()
	active := c.active
	c.mu.Unlock()
	return active
}

func (c *fakeConnection) SetReadTimeout(time.Duration) error   { return nil }
func (c *fakeConnection) SetWriteTimeout(time.Duration) error  { return nil }
func (c *fakeConnection) SetIdleTimeout(time.Duration) error   { return nil }
func (c *fakeConnection) SetOnRequest(netpoll.OnRequest) error { return nil }

func (c *fakeConnection) AddCloseCallback(callback netpoll.CloseCallback) error {
	c.mu.Lock()
	c.callbacks = append(c.callbacks, callback)
	c.mu.Unlock()
	return nil
}

func (c *fakeConnection) fireCloseCallbacks() {
	c.mu.Lock()
	if c.callbacksFired {
		c.mu.Unlock()
		return
	}
	c.callbacksFired = true
	callbacks := append([]netpoll.CloseCallback(nil), c.callbacks...)
	c.mu.Unlock()
	for i := len(callbacks) - 1; i >= 0; i-- {
		_ = callbacks[i](c)
	}
}

func (c *fakeConnection) writtenBytes() []byte {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]byte(nil), c.wire.Bytes()...)
}

func (c *fakeConnection) eventSequence() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]string(nil), c.events...)
}

type fakeNetpollWriter struct {
	conn    *fakeConnection
	pending []byte
}

func (w *fakeNetpollWriter) Malloc(n int) ([]byte, error) {
	start := len(w.pending)
	w.pending = append(w.pending, make([]byte, n)...)
	return w.pending[start:], nil
}

func (w *fakeNetpollWriter) WriteString(value string) (int, error) {
	w.pending = append(w.pending, value...)
	return len(value), nil
}

func (w *fakeNetpollWriter) WriteBinary(value []byte) (int, error) {
	w.pending = append(w.pending, value...)
	return len(value), nil
}

func (w *fakeNetpollWriter) WriteByte(value byte) error {
	w.pending = append(w.pending, value)
	return nil
}

func (w *fakeNetpollWriter) WriteDirect(value []byte, _ int) error {
	w.pending = append(w.pending, value...)
	return nil
}

func (w *fakeNetpollWriter) MallocAck(n int) error {
	if n < 0 || n > len(w.pending) {
		return errors.New("invalid malloc acknowledgment")
	}
	w.pending = w.pending[:n]
	return nil
}

func (w *fakeNetpollWriter) Append(netpoll.Writer) error {
	return errors.New("append is not supported by fake writer")
}

func (w *fakeNetpollWriter) Flush() error {
	w.conn.mu.Lock()
	_, err := w.conn.wire.Write(w.pending)
	w.conn.events = append(w.conn.events, "flush")
	w.pending = nil
	w.conn.mu.Unlock()
	return err
}

func (w *fakeNetpollWriter) MallocLen() int { return len(w.pending) }

type fakeAddr string

func (a fakeAddr) Network() string { return "tcp" }
func (a fakeAddr) String() string  { return string(a) }
