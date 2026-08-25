package nioserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/cloudwego/netpoll"
	"github.com/hujm2023/hlog"

	protocol "github.com/hujm2023/go-sms-protocol"
)

var (
	// ErrServerRunning indicates that the single-use server is already running.
	ErrServerRunning = errors.New("nioserver: server already running")
	// ErrServerStopped indicates that the single-use server has already stopped.
	ErrServerStopped = errors.New("nioserver: server stopped")
	// ErrServerShutdown is used as the cancellation cause during shutdown.
	ErrServerShutdown = errors.New("nioserver: server shutdown")
)

const (
	defaultMaxHandlers        = 1024
	defaultMaxHandlersPerConn = 1
	defaultMaxPendingWrites   = 64
	defaultCleanupTimeout     = 5 * time.Second
)

// UnpackFunc unpacks one complete PDU. A non-nil error is fatal for the
// connection; partial input must wait through the netpoll Reader API.
type UnpackFunc func(ctx context.Context, r netpoll.Reader) (protocol.PDU, error)

// HandleFunc handles one PDU and may return response bytes.
type HandleFunc func(ctx context.Context, p protocol.PDU) (respData []byte, err error)

// RefreshCtxFunc derives request metadata from the connection context.
type RefreshCtxFunc func(ctx context.Context) context.Context

// OnCloseFunc runs during phase-B cleanup, outside netpoll callbacks. The
// supplied context has an independent cleanup deadline; conn is already
// inactive and must only be used for identification or metadata reads.
type OnCloseFunc func(ctx context.Context, conn netpoll.Connection)

// ServerOption configures BaseServer.
type ServerOption[T any] func(*BaseServer[T])

// WithNetpollOptions adds transport options.
//
// Lifecycle callbacks supplied here are intentionally disabled by BaseServer;
// use WithOnCloseFunc and WithRefreshCtxWhenRead instead.
// Deprecated: use the explicit timeout options below.
func WithNetpollOptions[T any](opts ...netpoll.Option) ServerOption[T] {
	return func(s *BaseServer[T]) {
		s.options = append(s.options, opts...)
	}
}

func WithUnpackFunc[T any](unpack UnpackFunc) ServerOption[T] {
	return func(s *BaseServer[T]) {
		s.unpackBlock = unpack
	}
}

func WithHandleFunc[T any](handle HandleFunc) ServerOption[T] {
	return func(s *BaseServer[T]) {
		s.handle = handle
	}
}

func WithRefreshCtxWhenRead[T any](f RefreshCtxFunc) ServerOption[T] {
	return func(s *BaseServer[T]) {
		s.refreshCtxWhenRead = f
	}
}

func WithLogger[T any](logger hlog.FullLogger) ServerOption[T] {
	return func(s *BaseServer[T]) {
		s.logger = logger
	}
}

// WithWorkerPool supplies a caller-owned pool. BaseServer tracks tasks
// submitted to it but never closes or reconfigures the pool.
func WithWorkerPool[T any](pool gopool.Pool) ServerOption[T] {
	return func(s *BaseServer[T]) {
		s.workerpool = pool
	}
}

func WithOnCloseFunc[T any](f OnCloseFunc) ServerOption[T] {
	return func(s *BaseServer[T]) {
		s.closeFunc = f
	}
}

// WithMaxInFlightHandlers sets the global admitted handler limit. The permit
// covers time queued in the worker pool and time spent executing.
func WithMaxInFlightHandlers[T any](limit int) ServerOption[T] {
	return func(s *BaseServer[T]) {
		s.maxHandlers = limit
	}
}

// WithMaxInFlightHandlersPerConnection sets the per-connection admitted
// handler limit. Values above one opt into concurrent, potentially reordered
// responses; the caller remains responsible for sequence matching.
func WithMaxInFlightHandlersPerConnection[T any](limit int) ServerOption[T] {
	return func(s *BaseServer[T]) {
		s.maxHandlersPerConn = limit
	}
}

// WithMaxPendingWrites bounds active plus waiting writes per connection.
func WithMaxPendingWrites[T any](limit int) ServerOption[T] {
	return func(s *BaseServer[T]) {
		s.maxPendingWrites = limit
	}
}

// WithCleanupTimeout sets the independent phase-B cleanup deadline.
func WithCleanupTimeout[T any](timeout time.Duration) ServerOption[T] {
	return func(s *BaseServer[T]) {
		s.cleanupTimeout = timeout
	}
}

func WithReadTimeout[T any](timeout time.Duration) ServerOption[T] {
	return func(s *BaseServer[T]) {
		s.readTimeout = timeout
	}
}

func WithWriteTimeout[T any](timeout time.Duration) ServerOption[T] {
	return func(s *BaseServer[T]) {
		s.writeTimeout = timeout
	}
}

func WithIdleTimeout[T any](timeout time.Duration) ServerOption[T] {
	return func(s *BaseServer[T]) {
		s.idleTimeout = timeout
	}
}

type serverState uint8

const (
	serverStateUnknown serverState = iota
	serverStateNew
	serverStateServing
	serverStateShuttingDown
	serverStateStopped
)

// BaseServer is a single-use TCP server based on netpoll.
type BaseServer[T any] struct {
	network string
	address string
	options []netpoll.Option

	unpackBlock        UnpackFunc
	handle             HandleFunc
	refreshCtxWhenRead RefreshCtxFunc
	closeFunc          OnCloseFunc

	readTimeout  time.Duration
	writeTimeout time.Duration
	idleTimeout  time.Duration

	maxHandlers        int
	maxHandlersPerConn int
	maxPendingWrites   int
	cleanupTimeout     time.Duration

	listenerFactory func(network, address string) (netpoll.Listener, error)
	eventLoop       netpoll.EventLoop
	workerpool      gopool.Pool
	logger          hlog.FullLogger

	stateMu         sync.Mutex
	state           serverState
	listener        netpoll.Listener
	boundAddr       net.Addr
	runDone         chan struct{}
	runDoneClosed   bool
	runErr          error
	shutdownStarted bool
	shutdownDone    chan struct{}
	shutdownClosed  bool
	shutdownErr     error

	stopOnce sync.Once

	sessionsMu sync.Mutex
	accepting  bool
	sessions   map[*muxConn[T]]struct{}

	globalGate     *admissionGate
	cleanupTracker *taskTracker
}

// NewBaseServer validates configuration and builds the event loop. It does
// not bind a listener; binding belongs to Run.
func NewBaseServer[T any](network, address string, opts ...ServerOption[T]) (*BaseServer[T], error) {
	server := &BaseServer[T]{
		network:            network,
		address:            address,
		logger:             hlog.DefaultLogger(),
		maxHandlers:        defaultMaxHandlers,
		maxHandlersPerConn: defaultMaxHandlersPerConn,
		maxPendingWrites:   defaultMaxPendingWrites,
		cleanupTimeout:     defaultCleanupTimeout,
		listenerFactory:    netpoll.CreateListener,
		state:              serverStateNew,
		runDone:            make(chan struct{}),
		shutdownDone:       make(chan struct{}),
		sessions:           make(map[*muxConn[T]]struct{}),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(server)
		}
	}
	if server.unpackBlock == nil {
		return nil, errors.New("nioserver: unpack func is nil")
	}
	if server.handle == nil {
		return nil, errors.New("nioserver: handle func is nil")
	}
	if server.logger == nil {
		server.logger = hlog.DefaultLogger()
	}
	if server.maxHandlers < 1 {
		return nil, errors.New("nioserver: max in-flight handlers must be positive")
	}
	if server.maxHandlersPerConn < 1 {
		return nil, errors.New("nioserver: per-connection handler limit must be positive")
	}
	if server.maxPendingWrites < 1 {
		return nil, errors.New("nioserver: max pending writes must be positive")
	}
	if server.cleanupTimeout <= 0 {
		return nil, errors.New("nioserver: cleanup timeout must be positive")
	}
	if server.readTimeout < 0 || server.writeTimeout < 0 || server.idleTimeout < 0 {
		return nil, errors.New("nioserver: transport timeouts must not be negative")
	}

	dispatch := func(ctx context.Context, conn netpoll.Connection) error {
		err := server.DispatchRequest(ctx, conn)
		if err != nil && !errors.Is(err, context.Canceled) &&
			!errors.Is(err, ErrAdmissionClosed) && !errors.Is(err, ErrConnectionClosed) {
			server.logger.CtxErrorf(ctx, "dispatch request error: %v", err)
		}
		return err
	}

	transportOptions := append([]netpoll.Option(nil), server.options...)
	transportOptions = append(transportOptions,
		netpoll.WithReadTimeout(server.readTimeout),
		netpoll.WithWriteTimeout(server.writeTimeout),
		netpoll.WithIdleTimeout(server.idleTimeout),
		// Clear lifecycle callbacks that may have arrived through the deprecated
		// raw option escape hatch, then install the only close bridge owner.
		netpoll.WithOnConnect(nil),
		netpoll.WithOnDisconnect(nil),
		netpoll.WithOnPrepare(server.OnOpenConn),
	)
	eventLoop, err := netpoll.NewEventLoop(dispatch, transportOptions...)
	if err != nil {
		return nil, fmt.Errorf("nioserver: create event loop: %w", err)
	}
	server.eventLoop = eventLoop
	if server.workerpool == nil {
		server.workerpool = gopool.NewPool("nioserver", int32(server.maxHandlers), gopool.NewConfig())
	}
	server.globalGate = newAdmissionGate(server.maxHandlers)
	server.cleanupTracker = newTaskTracker()
	return server, nil
}

// Run binds the listener and synchronously serves until shutdown or an event
// loop error. It never installs process signal handlers and never panics for a
// serving error.
func (s *BaseServer[T]) Run() error {
	if err := s.beginRun(); err != nil {
		return err
	}

	listener, err := s.listenerFactory(s.network, s.address)
	if err != nil {
		runErr := fmt.Errorf("nioserver: create listener: %w", err)
		return s.finishRun(runErr)
	}

	s.stateMu.Lock()
	s.listener = listener
	s.boundAddr = listener.Addr()
	if s.shutdownStarted {
		s.stateMu.Unlock()
		_ = listener.Close()
		return s.finishRun(nil)
	}
	// Publish connection admission before releasing stateMu. Shutdown changes
	// shutdownStarted under the same lock and then closes admission, so it
	// cannot be lost between the startup check and this store.
	s.sessionsMu.Lock()
	s.accepting = true
	s.sessionsMu.Unlock()
	s.stateMu.Unlock()

	err = s.eventLoop.Serve(listener)
	_ = listener.Close()
	return s.finishRun(err)
}

func (s *BaseServer[T]) beginRun() error {
	s.stateMu.Lock()
	defer s.stateMu.Unlock()
	switch s.state {
	case serverStateNew:
		s.state = serverStateServing
		return nil
	case serverStateServing, serverStateShuttingDown:
		return ErrServerRunning
	default:
		return ErrServerStopped
	}
}

func (s *BaseServer[T]) finishRun(runErr error) error {
	s.stateMu.Lock()
	s.runErr = runErr
	s.closeRunDoneLocked()
	isOwner := !s.shutdownStarted
	if isOwner {
		s.shutdownStarted = true
		s.state = serverStateShuttingDown
	}
	shutdownDone := s.shutdownDone
	s.stateMu.Unlock()

	if isOwner {
		ctx, cancel := context.WithTimeout(context.Background(), s.cleanupTimeout)
		s.stopConnections(runErr)
		s.sealCleanupTracker()
		drainErr := s.waitForDrain(ctx)
		cancel()
		if drainErr != nil {
			s.logger.Errorf("server cleanup error: %v", drainErr)
		}
		s.finishShutdown(drainErr)
	} else {
		<-shutdownDone
	}
	return runErr
}

// Shutdown performs one context-bounded graceful shutdown. Concurrent callers
// wait for the same shutdown operation; an individual caller may stop waiting
// when its own context expires.
func (s *BaseServer[T]) Shutdown(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}

	s.stateMu.Lock()
	switch s.state {
	case serverStateNew:
		s.shutdownStarted = true
		s.state = serverStateShuttingDown
		s.stateMu.Unlock()
		s.stopConnections(ErrServerShutdown)
		s.cleanupTracker.Close()
		s.finishShutdown(nil)
		return nil
	case serverStateServing:
		if !s.shutdownStarted {
			s.shutdownStarted = true
			s.state = serverStateShuttingDown
			s.stateMu.Unlock()
			err := s.performShutdown(ctx)
			s.finishShutdown(err)
			return err
		}
	case serverStateShuttingDown:
		// Another caller owns the shutdown operation.
	case serverStateStopped:
		err := s.shutdownErr
		s.stateMu.Unlock()
		return err
	default:
		s.stateMu.Unlock()
		return ErrServerStopped
	}
	done := s.shutdownDone
	s.stateMu.Unlock()

	select {
	case <-done:
		s.stateMu.Lock()
		err := s.shutdownErr
		s.stateMu.Unlock()
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *BaseServer[T]) performShutdown(ctx context.Context) error {
	s.stopConnections(ErrServerShutdown)
	eventLoopErr := s.eventLoop.Shutdown(ctx)

	// Close also covers the narrow startup race where Shutdown reaches the
	// event loop before Serve has installed its internal server.
	s.stateMu.Lock()
	listener := s.listener
	s.stateMu.Unlock()
	if listener != nil {
		_ = listener.Close()
	}

	s.sealCleanupTracker()
	drainErr := s.waitForDrain(ctx)
	return errors.Join(eventLoopErr, drainErr)
}

func (s *BaseServer[T]) stopConnections(cause error) {
	s.stopOnce.Do(func() {
		s.globalGate.Close()

		s.sessionsMu.Lock()
		s.accepting = false
		sessions := make([]*muxConn[T], 0, len(s.sessions))
		for session := range s.sessions {
			sessions = append(sessions, session)
		}
		s.sessionsMu.Unlock()

		for _, session := range sessions {
			if err := session.closeWithCause(cause); err != nil && !errors.Is(err, net.ErrClosed) {
				s.logger.Errorf("close connection error: %v", err)
			}
		}
	})
}

func (s *BaseServer[T]) sealCleanupTracker() {
	// OnOpenConn registers or rejects sessions while holding sessionsMu. Taking
	// the same lock after EventLoop shutdown establishes the seal boundary.
	s.sessionsMu.Lock()
	s.cleanupTracker.Close()
	s.sessionsMu.Unlock()
}

func (s *BaseServer[T]) waitForDrain(ctx context.Context) error {
	handlerErr := s.globalGate.Wait(ctx)
	cleanupErr := s.cleanupTracker.Wait(ctx)
	return errors.Join(handlerErr, cleanupErr)
}

func (s *BaseServer[T]) finishShutdown(err error) {
	s.stateMu.Lock()
	s.shutdownErr = err
	s.state = serverStateStopped
	s.closeRunDoneLocked()
	if !s.shutdownClosed {
		close(s.shutdownDone)
		s.shutdownClosed = true
	}
	s.stateMu.Unlock()
}

func (s *BaseServer[T]) closeRunDoneLocked() {
	if !s.runDoneClosed {
		close(s.runDone)
		s.runDoneClosed = true
	}
}

// Addr returns nil until Run has successfully bound its listener.
func (s *BaseServer[T]) Addr() net.Addr {
	s.stateMu.Lock()
	addr := s.boundAddr
	s.stateMu.Unlock()
	return addr
}

// Serve retains the historical process-signal wrapper.
// Deprecated: applications should own signals and call Run and Shutdown.
func (s *BaseServer[T]) Serve(wait time.Duration) {
	runResult := make(chan error, 1)
	go func() {
		runResult <- s.Run()
	}()

	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-runResult:
		if err != nil {
			s.logger.Errorf("server run error: %v", err)
		}
	case <-signalCtx.Done():
		ctx, cancel := context.WithTimeout(context.Background(), wait)
		if err := s.Shutdown(ctx); err != nil {
			s.logger.Errorf("server shutdown error: %v", err)
		}
		cancel()
		if err := <-runResult; err != nil {
			s.logger.Errorf("server run error: %v", err)
		}
	}
}

// DispatchRequest unpacks one PDU, acquires bounded handler admission, and
// submits it to the configured worker pool.
func (s *BaseServer[T]) DispatchRequest(ctx context.Context, conn netpoll.Connection) error {
	mc, ok := sessionFromContext[T](ctx)
	if !ok {
		_ = conn.Close()
		return errors.New("nioserver: connection missing from context")
	}

	if s.refreshCtxWhenRead != nil {
		ctx = s.refreshCtxWhenRead(ctx)
		if ctx == nil {
			_ = mc.closeWithCause(ErrConnectionClosed)
			return errors.New("nioserver: refreshed context is nil")
		}
		ctx = fillCtx[T](ctx, mc)
	}

	pdu, err := s.unpackBlock(ctx, conn.Reader())
	if err != nil {
		wrapped := fmt.Errorf("unpack pdu: %w", err)
		_ = mc.closeWithCause(wrapped)
		return wrapped
	}

	if err := mc.handlerGate.Acquire(ctx); err != nil {
		_ = mc.closeWithCause(err)
		return err
	}
	if err := s.globalGate.Acquire(ctx); err != nil {
		mc.handlerGate.Release()
		_ = mc.closeWithCause(err)
		return err
	}

	task := func() {
		defer s.globalGate.Release()
		defer mc.handlerGate.Release()
		defer func() {
			if recovered := recover(); recovered != nil {
				s.logger.CtxErrorf(ctx, "handle pdu panic: %v\n%s", recovered, debug.Stack())
				_ = mc.closeWithCause(fmt.Errorf("handle pdu panic: %v", recovered))
			}
		}()
		s.handlePDU(ctx, mc, pdu)
	}
	if err := submitPoolTask(s.workerpool, ctx, task); err != nil {
		s.globalGate.Release()
		mc.handlerGate.Release()
		_ = mc.closeWithCause(err)
		return err
	}
	return nil
}

func (s *BaseServer[T]) handlePDU(ctx context.Context, mc *muxConn[T], pdu protocol.PDU) {
	response, handleErr := s.handle(ctx, pdu)
	if handleErr != nil {
		s.logger.CtxErrorf(ctx, "handle pdu error: %v", handleErr)
		if len(response) > 0 {
			if err := mc.WriteAndClose(ctx, response); err != nil {
				s.logger.CtxErrorf(ctx, "write response before close error: %v", err)
			}
			return
		}
		_ = mc.closeWithCause(handleErr)
		return
	}
	if len(response) > 0 {
		if err := mc.Write(ctx, response); err != nil {
			s.logger.CtxErrorf(ctx, "write response error: %v", err)
			_ = mc.closeWithCause(err)
		}
	}
}

func submitPoolTask(pool gopool.Pool, ctx context.Context, task func()) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("nioserver: submit worker task panic: %v", recovered)
		}
	}()
	pool.CtxGo(ctx, task)
	return nil
}

// OnOpenConn installs the single early CloseCallback bridge before exposing
// the session to request handling.
func (s *BaseServer[T]) OnOpenConn(conn netpoll.Connection) context.Context {
	mc := newSvrMuxConn(conn, s)
	ctx := fillCtx[T](mc.connectionContext(), mc)

	err := conn.AddCloseCallback(func(netpoll.Connection) error {
		mc.transportClosed(ErrConnectionClosed)
		return nil
	})
	if err != nil {
		cause := fmt.Errorf("register close callback: %w", err)
		s.logger.CtxErrorf(ctx, "register close callback error: %v", err)
		_ = mc.closeWithCause(cause)
		return ctx
	}

	s.sessionsMu.Lock()
	if s.accepting {
		s.sessions[mc] = struct{}{}
		s.sessionsMu.Unlock()
		s.logger.Noticef("[OnOpenConn] %s connected", mc.RemoteAddr())
		return ctx
	}

	// A connection prepared across the shutdown boundary is closed before it
	// can be exposed. Holding sessionsMu keeps cleanup scheduling before the
	// shutdown tracker seal.
	_ = mc.closeWithCause(ErrServerShutdown)
	s.sessionsMu.Unlock()
	return ctx
}

// OnCloseConn is retained for callers that invoked it directly. BaseServer no
// longer registers it as netpoll OnDisconnect.
func (s *BaseServer[T]) OnCloseConn(ctx context.Context, _ netpoll.Connection) {
	if mc, ok := sessionFromContext[T](ctx); ok {
		mc.transportClosed(ErrConnectionClosed)
	}
}

func sessionFromContext[T any](ctx context.Context) (*muxConn[T], bool) {
	conn, ok := GetCtxConn[T](ctx)
	if !ok {
		return nil, false
	}
	mc, ok := conn.(*muxConn[T])
	return mc, ok
}

func (s *BaseServer[T]) scheduleCleanup(mc *muxConn[T]) {
	if !s.cleanupTracker.TryAdd() {
		// This indicates a connection crossed the event-loop shutdown boundary.
		// Phase A is still complete; do not run user code on the netpoll callback.
		s.logger.Errorf("connection cleanup scheduled after server stopped: %s", mc.RemoteAddr())
		close(mc.cleanupDone)
		return
	}
	go func() {
		defer s.cleanupTracker.Done()
		defer close(mc.cleanupDone)
		defer s.removeSession(mc)

		cleanupCtx, cancel := context.WithTimeout(context.Background(), s.cleanupTimeout)
		cleanupCtx = fillCtx[T](cleanupCtx, mc)
		defer cancel()

		if err := mc.handlerGate.Wait(cleanupCtx); err != nil {
			s.logger.CtxErrorf(cleanupCtx, "wait connection handlers error: %v", err)
		}
		if err := mc.writer.Stop(cleanupCtx); err != nil {
			s.logger.CtxErrorf(cleanupCtx, "stop connection writer error: %v", err)
		}

		if s.closeFunc != nil {
			func() {
				defer func() {
					if recovered := recover(); recovered != nil {
						s.logger.CtxErrorf(cleanupCtx, "connection close callback panic: %v\n%s", recovered, debug.Stack())
					}
				}()
				s.closeFunc(cleanupCtx, mc.conn)
			}()
		}
		s.logger.Noticef("[OnCloseConn] %s closed", mc.RemoteAddr())
	}()
}

func (s *BaseServer[T]) removeSession(mc *muxConn[T]) {
	s.sessionsMu.Lock()
	delete(s.sessions, mc)
	s.sessionsMu.Unlock()
}

// taskTracker tracks unbounded-but-finite cleanup tasks. Close seals new task
// admission; Wait is context-aware and uses change notifications, not polling.
type taskTracker struct {
	mu      sync.Mutex
	closed  bool
	count   int
	changed chan struct{}
}

func newTaskTracker() *taskTracker {
	return &taskTracker{changed: make(chan struct{})}
}

func (t *taskTracker) TryAdd() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return false
	}
	t.count++
	t.signalLocked()
	return true
}

func (t *taskTracker) Done() {
	t.mu.Lock()
	if t.count == 0 {
		t.mu.Unlock()
		panic("nioserver: cleanup tracker done without task")
	}
	t.count--
	t.signalLocked()
	t.mu.Unlock()
}

func (t *taskTracker) Close() {
	t.mu.Lock()
	if !t.closed {
		t.closed = true
		t.signalLocked()
	}
	t.mu.Unlock()
}

func (t *taskTracker) Wait(ctx context.Context) error {
	for {
		t.mu.Lock()
		if t.closed && t.count == 0 {
			t.mu.Unlock()
			return nil
		}
		changed := t.changed
		t.mu.Unlock()

		select {
		case <-changed:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (t *taskTracker) signalLocked() {
	close(t.changed)
	t.changed = make(chan struct{})
}
