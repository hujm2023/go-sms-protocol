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

type (
	UnpackFunc     func(ctx context.Context, r netpoll.Reader) (protocol.PDU, error)
	HandleFunc     func(ctx context.Context, p protocol.PDU) (respData []byte, err error)
	RefreshCtxFunc func(ctx context.Context) context.Context
	OnCloseFunc    func(ctx context.Context, conn netpoll.Connection)
)

type ServerOption[T any] func(*BaseServer[T])

func WithNetpollOptions[T any](opts ...netpoll.Option) ServerOption[T] {
	return func(s *BaseServer[T]) {
		s.options = append(s.options, opts...)
	}
}

func WithUnpackFunc[T any](unpack UnpackFunc) ServerOption[T] {
	return func(s *BaseServer[T]) {
		s.unpack = unpack
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

// 修改 WithWorkerPool 以支持泛型 T
func WithWorkerPool[T any](pool gopool.Pool) ServerOption[T] {
	return func(s *BaseServer[T]) {
		s.workerpool = pool
	}
}

// 修改 WithOnCloseFunc 以支持泛型 T
func WithOnCloseFunc[T any](f OnCloseFunc) ServerOption[T] {
	return func(s *BaseServer[T]) {
		s.closeFunc = f
	}
}

// 修改 BaseServer 以支持泛型 T
type BaseServer[T any] struct {
	network, address string
	options          []netpoll.Option

	unpack             UnpackFunc
	handle             HandleFunc
	refreshCtxWhenRead RefreshCtxFunc
	closeFunc          OnCloseFunc

	listener  netpoll.Listener
	eventLoop netpoll.EventLoop

	workerpool gopool.Pool
	logger     hlog.FullLogger
}

// 修改 NewBaseServer 以支持泛型 T
func NewBaseServer[T any](network, address string, opts ...ServerOption[T]) (*BaseServer[T], error) {
	server := &BaseServer[T]{
		network: network,
		address: address,
	}
	for _, opt := range opts {
		opt(server)
	}
	if server.unpack == nil {
		return nil, fmt.Errorf("unpack func is nil")
	}
	if server.handle == nil {
		return nil, fmt.Errorf("handle func is nil")
	}

	// 注意：这里需要将 server 的方法绑定传递给 netpoll
	// 由于 Go 泛型方法的限制，我们需要创建闭包
	onPrepare := func(conn netpoll.Connection) context.Context {
		return server.OnOpenConn(conn)
	}
	onDisconnect := func(ctx context.Context, conn netpoll.Connection) {
		server.OnCloseConn(ctx, conn)
	}
	dispatch := func(ctx context.Context, conn netpoll.Connection) error {
		return server.DispatchRequest(ctx, conn)
	}

	nOpts := []netpoll.Option{
		netpoll.WithOnPrepare(onPrepare),
		netpoll.WithOnDisconnect(onDisconnect),
	}
	server.options = append(server.options, nOpts...)
	eventLoop, err := netpoll.NewEventLoop(
		dispatch, // 使用闭包
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
	if server.logger == nil {
		server.logger = hlog.DefaultLogger()
	}

	return server, nil
}

// 修改 Serve 的接收者
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

// 修改 DispatchRequest 以支持泛型 T
func (s *BaseServer[T]) DispatchRequest(ctx context.Context, conn netpoll.Connection) error {
	if s.refreshCtxWhenRead != nil {
		ctx = s.refreshCtxWhenRead(ctx)
	}
	// 从ctx中解析出自己的 conn
	// 使用泛型 GetCtxConn
	mc, ok := GetCtxConn[T](ctx)
	if !ok {
		// 理论上不应该发生，因为 OnOpenConn 总是会填充 context
		s.logger.CtxErrorf(ctx, "failed to get connection from context")
		_ = conn.Close() // 关闭连接以防万一
		return fmt.Errorf("failed to get connection from context")
	}

	reader := conn.Reader()

	// unpack中调用
	// 1. 这里是水平触发的，只读一个包就行。如果没有具体的数据，建议通过 Next(n) 等待
	// 2. 如果使用了 Zero-Copy 相关函数，记得 Release
	rPDU, err := s.unpack(ctx, reader)
	if err != nil {
		return err
	}

	// 这里必须使用异步goroutine去处理，不能阻塞整个eventloop
	// 在这个goroutine中，不应该在对conn进行读写
	s.workerpool.Go(func() {
		// 处理这个PDU
		respData, err := s.handle(ctx, rPDU)
		if err != nil {
			s.logger.CtxErrorf(ctx, "handle pdu error: %v", err)
			// 关闭连接
			_ = mc.(interface{ Close() error }).Close() // 需要类型断言来调用 Close
			return
		}
		if len(respData) == 0 {
			// 没有要回写的数据
			return
		}

		s.logger.CtxDebugf(ctx, " == write data: %+v", respData)
		mc.AsyncWrite(ctx, respData)
	})

	return nil
}

// 修改 OnOpenConn 以支持泛型 T
func (s *BaseServer[T]) OnOpenConn(conn netpoll.Connection) context.Context {
	s.logger.Noticef("[OnOpenConn] %s connected", conn.RemoteAddr().String())
	// 创建泛型 muxConn
	mc := newSvrMuxConn[T](conn)
	// 使用泛型 fillCtx
	ctx := fillCtx[T](context.Background(), mc)
	return ctx
}

// 修改 OnCloseConn 以支持泛型 T
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
	} else {
		s.logger.CtxErrorf(ctx, "connection is not of type *muxConn[T] on close")
	}
}
