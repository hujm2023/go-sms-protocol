package nioserver

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/cloudwego/netpoll"
	"github.com/hujm2023/hlog"

	protocol "github.com/hujm2023/go-sms-protocol"
)

// UnpackFunc defines the function signature for unpacking data from the reader into a PDU.
type UnpackFunc func(ctx context.Context, r netpoll.Reader) (protocol.PDU, error)

// HandleFunc defines the function signature for handling business logic for a PDU.
type HandleFunc func(ctx context.Context, p protocol.PDU) (respData []byte, err error)

// RefreshCtxFunc defines the function signature for refreshing context before reading data.
type RefreshCtxFunc func(ctx context.Context) context.Context

// OnCloseFunc defines the function signature for the callback when a connection is closed.
type OnCloseFunc func(ctx context.Context, conn netpoll.Connection)

// ServerOption is the function option type for configuring BaseServer.
type ServerOption[T any] func(*BaseServer[T])

// WithNetpollOptions adds underlying netpoll options to the BaseServer.
func WithNetpollOptions[T any](opts ...netpoll.Option) ServerOption[T] {
	return func(s *BaseServer[T]) {
		s.options = append(s.options, opts...)
	}
}

// WithUnpackFunc sets the unpack function for the BaseServer.
func WithUnpackFunc[T any](unpack UnpackFunc) ServerOption[T] {
	return func(s *BaseServer[T]) {
		s.unpackBlock = unpack
	}
}

// WithHandleFunc sets the business handler function for the BaseServer.
func WithHandleFunc[T any](handle HandleFunc) ServerOption[T] {
	return func(s *BaseServer[T]) {
		s.handle = handle
	}
}

// WithRefreshCtxWhenRead sets the context refresh function for the BaseServer.
func WithRefreshCtxWhenRead[T any](f RefreshCtxFunc) ServerOption[T] {
	return func(s *BaseServer[T]) {
		s.refreshCtxWhenRead = f
	}
}

// WithLogger sets the logger for the BaseServer.
func WithLogger[T any](logger hlog.FullLogger) ServerOption[T] {
	return func(s *BaseServer[T]) {
		s.logger = logger
	}
}

// WithWorkerPool sets the worker goroutine pool for the BaseServer.
func WithWorkerPool[T any](pool gopool.Pool) ServerOption[T] {
	return func(s *BaseServer[T]) {
		s.workerpool = pool
	}
}

// WithOnCloseFunc sets the connection close callback function.
func WithOnCloseFunc[T any](f OnCloseFunc) ServerOption[T] {
	return func(s *BaseServer[T]) {
		s.closeFunc = f
	}
}

// BaseServer is a generic TCP server implementation based on netpoll.
// It handles connection management, data unpacking, business logic dispatching, and graceful shutdown.
type BaseServer[T any] struct {
	network, address string           // Network type and address to listen on
	options          []netpoll.Option // Netpoll configuration options

	unpackBlock        UnpackFunc     // Data unpack function
	handle             HandleFunc     // Business handler function
	refreshCtxWhenRead RefreshCtxFunc // Context refresh function before read
	closeFunc          OnCloseFunc    // Connection close callback

	listener  netpoll.Listener  // Network listener
	eventLoop netpoll.EventLoop // Netpoll event loop

	workerpool gopool.Pool     // Goroutine pool for business logic
	logger     hlog.FullLogger // Logger instance
}

// NewBaseServer creates and initializes a new BaseServer instance.
// Requires network type, address, and server options. UnpackFunc and HandleFunc are mandatory.
func NewBaseServer[T any](network, address string, opts ...ServerOption[T]) (*BaseServer[T], error) {
	server := &BaseServer[T]{
		network: network,
		address: address,
		logger:  hlog.DefaultLogger(),
	}
	for _, opt := range opts {
		opt(server)
	}
	if server.unpackBlock == nil {
		return nil, fmt.Errorf("unpack func is nil")
	}
	if server.handle == nil {
		return nil, fmt.Errorf("handle func is nil")
	}

	onPrepare := func(conn netpoll.Connection) context.Context {
		return server.OnOpenConn(conn)
	}
	onDisconnect := func(ctx context.Context, conn netpoll.Connection) {
		server.OnCloseConn(ctx, conn)
	}
	dispatch := func(ctx context.Context, conn netpoll.Connection) error {
		err := server.DispatchRequest(ctx, conn)
		if err != nil {
			// For netpoll, the error returned here will be ignored.
			// If the connection is not closed here, OnRequest will be continuously triggered, resulting in an infinite loop
			server.logger.CtxErrorf(ctx, "dispatch request error: %v", err)
			_ = conn.Close()
		}
		return err
	}

	nOpts := []netpoll.Option{
		netpoll.WithOnPrepare(onPrepare),
		netpoll.WithOnDisconnect(onDisconnect),
	}
	server.options = append(server.options, nOpts...)
	eventLoop, err := netpoll.NewEventLoop(
		dispatch,
		server.options...,
	)
	if err != nil {
		return nil, fmt.Errorf("new event loop error: %w", err)
	}

	listener, err := netpoll.CreateListener(server.network, server.address)
	if err != nil {
		return nil, fmt.Errorf("create listener error: %w", err)
	}
	server.listener = listener
	server.eventLoop = eventLoop
	if server.workerpool == nil {
		server.workerpool = gopool.NewPool(
			"default",
			1024,
			gopool.NewConfig(),
		)
	}

	return server, nil
}

// Serve starts the server's event loop and begins accepting connections.
// It blocks until a termination signal (SIGINT, SIGTERM) is received or an error occurs.
// 'wait' specifies the timeout duration for graceful shutdown.
func (s *BaseServer[T]) Serve(wait time.Duration) {
	errChan := make(chan error)
	go func() {
		errChan <- s.eventLoop.Serve(s.listener)
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	select {
	case <-sig:
		s.logger.Noticef("received signal, exiting...")
		if err := s.eventLoop.Shutdown(ctx); err != nil {
			s.logger.Errorf("shutdown error: %v", err)
			return
		}
	case err := <-errChan:
		if err != nil {
			panic(err)
		}
	}
	s.logger.Notice("!!!exited done.")
}

// DispatchRequest is the netpoll request dispatch callback.
// It reads data, unpacks it using UnpackFunc, and submits the PDU to the worker pool for handling by HandleFunc.
func (s *BaseServer[T]) DispatchRequest(ctx context.Context, conn netpoll.Connection) error {
	if s.refreshCtxWhenRead != nil {
		ctx = s.refreshCtxWhenRead(ctx)
	}
	// From ctx get the own conn
	// Use generic GetCtxConn
	mc, ok := GetCtxConn[T](ctx)
	if !ok {
		// Theoretically should not happen, because OnOpenConn always fills context
		s.logger.CtxErrorf(ctx, "failed to get connection from context")
		_ = conn.Close() // Close connection just in case
		return fmt.Errorf("failed to get connection from context")
	}

	reader := conn.Reader()

	// unpack invoke
	// 1. 这里是水平触发的，只读一个包就行。如果没有具体的数据，建议通过 Next(n) 等待
	// 2. 如果使用了 Zero-Copy 相关函数，记得 Release
	rPDU, err := s.unpackBlock(ctx, reader)
	if err != nil {
		return err
	}

	// 这里必须使用异步goroutine去处理，不能阻塞整个eventloop
	// 在这个goroutine中，不应该在对conn进行读写
	s.workerpool.Go(func() {
		// 处理这个PDU
		respData, err := s.handle(ctx, rPDU)
		// 有数据要返回，先写，再判断是否要关闭
		if len(respData) > 0 {
			// s.logger.CtxDebugf(ctx, " == write data: %+v", respData)
			mc.AsyncWrite(ctx, respData)
		}

		if err != nil {
			s.logger.CtxErrorf(ctx, "handle pdu error: %v", err)
			// 关闭连接
			_ = mc.Close()
			return
		}
	})

	return nil
}

// OnOpenConn is the netpoll connection established callback (set via WithOnPrepare).
// It creates a muxConn instance and populates the context with it.
func (s *BaseServer[T]) OnOpenConn(conn netpoll.Connection) context.Context {
	s.logger.Noticef("[OnOpenConn] %s connected", conn.RemoteAddr().String())
	// 创建泛型 muxConn
	mc := newSvrMuxConn[T](conn)
	// 使用泛型 fillCtx
	ctx := fillCtx[T](context.Background(), mc)
	return ctx
}

// OnCloseConn is the netpoll connection closed callback (set via WithOnDisconnect).
// It logs the closure, executes the user-defined OnCloseFunc, and closes associated resources.
func (s *BaseServer[T]) OnCloseConn(ctx context.Context, conn netpoll.Connection) {
	s.logger.Noticef("[OnCloseConn] %s closed", conn.RemoteAddr().String())

	if s.closeFunc != nil {
		s.closeFunc(ctx, conn)
	}

	// 使用泛型 GetCtxConn
	mc, ok := GetCtxConn[T](ctx)
	if !ok {
		s.logger.CtxErrorf(ctx, "failed to get connection from context on close")
		return
	}

	// 需要类型断言来访问内部的 wqueue
	if muxConnInst, ok := mc.(*muxConn[T]); ok {
		_ = muxConnInst.wqueue.Close()
	}
}
