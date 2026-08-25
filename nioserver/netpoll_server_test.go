package nioserver

import (
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

func TestBaseServerRunBindsLazilyAndPropagatesEventLoopError(t *testing.T) {
	wantErr := errors.New("serve failed")
	eventLoop := newFakeEventLoop(wantErr)
	listener := newFakeListener("127.0.0.1:12345")
	server := newConnectionTestServer[any](t)
	var factoryCalls atomic.Int32
	server.eventLoop = eventLoop
	server.listenerFactory = func(string, string) (netpoll.Listener, error) {
		factoryCalls.Add(1)
		return listener, nil
	}

	if got := server.Addr(); got != nil {
		t.Fatalf("Addr() before Run = %v, want nil", got)
	}
	if got := factoryCalls.Load(); got != 0 {
		t.Fatalf("listener factory calls before Run = %d, want 0", got)
	}

	runResult := make(chan error, 1)
	go func() { runResult <- server.Run() }()
	waitForChannel(t, eventLoop.serveStarted, "event loop Serve")
	if got := server.Addr(); got == nil || got.String() != "127.0.0.1:12345" {
		t.Fatalf("Addr() after bind = %v, want 127.0.0.1:12345", got)
	}
	eventLoop.finishServe()
	if err := receiveError(t, runResult, "Run result"); !errors.Is(err, wantErr) {
		t.Fatalf("Run() error = %v, want %v", err, wantErr)
	}
	if err := server.Run(); !errors.Is(err, ErrServerStopped) {
		t.Fatalf("second Run() error = %v, want ErrServerStopped", err)
	}
}

func TestBaseServerShutdownWaitsForAdmittedHandlerAndIsIdempotent(t *testing.T) {
	eventLoop := newFakeEventLoop(nil)
	server := newConnectionTestServer[any](t)
	server.eventLoop = eventLoop
	server.listenerFactory = func(string, string) (netpoll.Listener, error) {
		return newFakeListener("127.0.0.1:12346"), nil
	}

	runResult := make(chan error, 1)
	go func() { runResult <- server.Run() }()
	waitForChannel(t, eventLoop.serveStarted, "event loop Serve")
	if err := server.globalGate.Acquire(context.Background()); err != nil {
		t.Fatalf("global handler Acquire() error = %v", err)
	}

	shutdownResult := make(chan error, 1)
	go func() { shutdownResult <- server.Shutdown(context.Background()) }()
	waitForChannel(t, eventLoop.shutdownStarted, "event loop Shutdown")
	select {
	case err := <-shutdownResult:
		t.Fatalf("Shutdown returned before admitted handler drained: %v", err)
	default:
	}

	server.globalGate.Release()
	if err := receiveError(t, shutdownResult, "Shutdown result"); err != nil {
		t.Fatalf("Shutdown() error = %v", err)
	}
	if err := receiveError(t, runResult, "Run result"); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if err := server.Shutdown(context.Background()); err != nil {
		t.Fatalf("repeated Shutdown() error = %v", err)
	}
}

func TestBaseServerShutdownContextStopsWaitingAndPersistsResult(t *testing.T) {
	eventLoop := newFakeEventLoop(nil)
	server := newConnectionTestServer[any](t)
	server.eventLoop = eventLoop
	server.listenerFactory = func(string, string) (netpoll.Listener, error) {
		return newFakeListener("127.0.0.1:12347"), nil
	}

	runResult := make(chan error, 1)
	go func() { runResult <- server.Run() }()
	waitForChannel(t, eventLoop.serveStarted, "event loop Serve")
	if err := server.globalGate.Acquire(context.Background()); err != nil {
		t.Fatalf("global handler Acquire() error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	shutdownResult := make(chan error, 1)
	go func() { shutdownResult <- server.Shutdown(ctx) }()
	waitForChannel(t, eventLoop.shutdownStarted, "event loop Shutdown")
	cancel()
	if err := receiveError(t, shutdownResult, "canceled Shutdown result"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Shutdown() error = %v, want context.Canceled", err)
	}
	if err := receiveError(t, runResult, "Run result"); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if err := server.Shutdown(context.Background()); !errors.Is(err, context.Canceled) {
		t.Fatalf("repeated Shutdown() error = %v, want stored context.Canceled", err)
	}
	server.globalGate.Release()
}

func TestBaseServerShutdownBeforeRunStopsSingleUseServer(t *testing.T) {
	server := newConnectionTestServer[any](t)
	if err := server.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() before Run error = %v", err)
	}
	if err := server.Run(); !errors.Is(err, ErrServerStopped) {
		t.Fatalf("Run() after Shutdown error = %v, want ErrServerStopped", err)
	}
	if server.Addr() != nil {
		t.Fatalf("Addr() after Shutdown-before-Run = %v, want nil", server.Addr())
	}
}

func TestBaseServerRefreshContextPreservesSession(t *testing.T) {
	type metadataKey struct{}
	handled := make(chan error, 1)
	server, err := NewBaseServer[any](
		"tcp",
		"127.0.0.1:0",
		WithUnpackFunc[any](func(context.Context, netpoll.Reader) (protocol.PDU, error) {
			return nil, nil
		}),
		WithRefreshCtxWhenRead[any](func(ctx context.Context) context.Context {
			return context.WithValue(ctx, metadataKey{}, "request metadata")
		}),
		WithHandleFunc[any](func(ctx context.Context, _ protocol.PDU) ([]byte, error) {
			if ctx.Value(metadataKey{}) != "request metadata" {
				handled <- errors.New("request metadata missing")
				return nil, nil
			}
			if _, ok := GetCtxConn[any](ctx); !ok {
				handled <- errors.New("connection session missing")
				return nil, nil
			}
			handled <- nil
			return nil, nil
		}),
	)
	if err != nil {
		t.Fatalf("NewBaseServer() error = %v", err)
	}
	setServerAccepting(server)
	conn := newFakeConnection()
	ctx := server.OnOpenConn(conn)
	if err := server.DispatchRequest(ctx, conn); err != nil {
		t.Fatalf("DispatchRequest() error = %v", err)
	}
	if err := receiveError(t, handled, "handler result"); err != nil {
		t.Fatal(err)
	}
	mc, _ := sessionFromContext[any](ctx)
	_ = mc.Close()
	waitForChannel(t, mc.cleanupDone, "connection cleanup")
	server.globalGate.Close()
	server.cleanupTracker.Close()
}

func TestBaseServerDefaultConnectionAdmissionPreservesHandlerOrder(t *testing.T) {
	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	secondFinished := make(chan struct{})
	var calls atomic.Int32
	var active atomic.Int32
	var maxActive atomic.Int32

	server, err := NewBaseServer[any](
		"tcp",
		"127.0.0.1:0",
		WithUnpackFunc[any](func(context.Context, netpoll.Reader) (protocol.PDU, error) {
			return nil, nil
		}),
		WithHandleFunc[any](func(context.Context, protocol.PDU) ([]byte, error) {
			call := calls.Add(1)
			current := active.Add(1)
			for {
				maximum := maxActive.Load()
				if current <= maximum || maxActive.CompareAndSwap(maximum, current) {
					break
				}
			}
			defer active.Add(-1)
			if call == 1 {
				close(firstStarted)
				<-releaseFirst
			} else {
				close(secondFinished)
			}
			return nil, nil
		}),
	)
	if err != nil {
		t.Fatalf("NewBaseServer() error = %v", err)
	}
	setServerAccepting(server)
	conn := newFakeConnection()
	ctx := server.OnOpenConn(conn)
	if err := server.DispatchRequest(ctx, conn); err != nil {
		t.Fatalf("first DispatchRequest() error = %v", err)
	}
	waitForChannel(t, firstStarted, "first handler")

	secondDispatch := make(chan error, 1)
	go func() { secondDispatch <- server.DispatchRequest(ctx, conn) }()
	select {
	case err := <-secondDispatch:
		t.Fatalf("second DispatchRequest returned before first handler completed: %v", err)
	default:
	}
	close(releaseFirst)
	if err := receiveError(t, secondDispatch, "second DispatchRequest result"); err != nil {
		t.Fatalf("second DispatchRequest() error = %v", err)
	}
	waitForChannel(t, secondFinished, "second handler")
	if got := maxActive.Load(); got != 1 {
		t.Fatalf("maximum concurrent handlers = %d, want 1", got)
	}

	mc, _ := sessionFromContext[any](ctx)
	_ = mc.Close()
	waitForChannel(t, mc.cleanupDone, "connection cleanup")
	server.globalGate.Close()
	server.cleanupTracker.Close()
}

func TestBaseServerGlobalAdmissionBoundsConnections(t *testing.T) {
	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	secondStarted := make(chan struct{})
	var calls atomic.Int32
	server, err := NewBaseServer[any](
		"tcp",
		"127.0.0.1:0",
		WithMaxInFlightHandlers[any](1),
		WithUnpackFunc[any](func(context.Context, netpoll.Reader) (protocol.PDU, error) {
			return nil, nil
		}),
		WithHandleFunc[any](func(context.Context, protocol.PDU) ([]byte, error) {
			if calls.Add(1) == 1 {
				close(firstStarted)
				<-releaseFirst
			} else {
				close(secondStarted)
			}
			return nil, nil
		}),
	)
	if err != nil {
		t.Fatalf("NewBaseServer() error = %v", err)
	}
	setServerAccepting(server)
	firstConn := newFakeConnection()
	firstCtx := server.OnOpenConn(firstConn)
	secondConn := newFakeConnection()
	secondCtx := server.OnOpenConn(secondConn)
	if err := server.DispatchRequest(firstCtx, firstConn); err != nil {
		t.Fatalf("first DispatchRequest() error = %v", err)
	}
	waitForChannel(t, firstStarted, "first global handler")

	secondDispatch := make(chan error, 1)
	go func() { secondDispatch <- server.DispatchRequest(secondCtx, secondConn) }()
	select {
	case err := <-secondDispatch:
		t.Fatalf("second connection crossed global limit: %v", err)
	default:
	}
	if got := server.globalGate.InFlight(); got != 1 {
		t.Fatalf("global in-flight handlers = %d, want 1", got)
	}
	close(releaseFirst)
	if err := receiveError(t, secondDispatch, "second global DispatchRequest"); err != nil {
		t.Fatalf("second DispatchRequest() error = %v", err)
	}
	waitForChannel(t, secondStarted, "second global handler")

	firstSession, _ := sessionFromContext[any](firstCtx)
	secondSession, _ := sessionFromContext[any](secondCtx)
	_ = firstSession.Close()
	_ = secondSession.Close()
	waitForChannel(t, firstSession.cleanupDone, "first connection cleanup")
	waitForChannel(t, secondSession.cleanupDone, "second connection cleanup")
	server.globalGate.Close()
	server.cleanupTracker.Close()
}

func TestBaseServerExplicitConnectionConcurrencyAllowsParallelHandlers(t *testing.T) {
	entered := make(chan struct{}, 2)
	release := make(chan struct{})
	server, err := NewBaseServer[any](
		"tcp",
		"127.0.0.1:0",
		WithMaxInFlightHandlersPerConnection[any](2),
		WithUnpackFunc[any](func(context.Context, netpoll.Reader) (protocol.PDU, error) {
			return nil, nil
		}),
		WithHandleFunc[any](func(context.Context, protocol.PDU) ([]byte, error) {
			entered <- struct{}{}
			<-release
			return nil, nil
		}),
	)
	if err != nil {
		t.Fatalf("NewBaseServer() error = %v", err)
	}
	setServerAccepting(server)
	conn := newFakeConnection()
	ctx := server.OnOpenConn(conn)
	if err := server.DispatchRequest(ctx, conn); err != nil {
		t.Fatalf("first DispatchRequest() error = %v", err)
	}
	if err := server.DispatchRequest(ctx, conn); err != nil {
		t.Fatalf("second DispatchRequest() error = %v", err)
	}
	waitForChannel(t, entered, "first concurrent handler")
	waitForChannel(t, entered, "second concurrent handler")

	close(release)
	mc, _ := sessionFromContext[any](ctx)
	_ = mc.Close()
	waitForChannel(t, mc.cleanupDone, "parallel handler cleanup")
	server.globalGate.Close()
	server.cleanupTracker.Close()
}

func TestBaseServerUnpackErrorIsFatalAndCleanupRunsOnce(t *testing.T) {
	wantErr := errors.New("invalid frame")
	callbackCount := atomic.Int32{}
	server, err := NewBaseServer[any](
		"tcp",
		"127.0.0.1:0",
		WithUnpackFunc[any](func(context.Context, netpoll.Reader) (protocol.PDU, error) {
			return nil, wantErr
		}),
		WithHandleFunc[any](func(context.Context, protocol.PDU) ([]byte, error) {
			t.Fatal("handler called after unpack error")
			return nil, nil
		}),
		WithOnCloseFunc[any](func(context.Context, netpoll.Connection) {
			callbackCount.Add(1)
		}),
	)
	if err != nil {
		t.Fatalf("NewBaseServer() error = %v", err)
	}
	setServerAccepting(server)
	conn := newFakeConnection()
	ctx := server.OnOpenConn(conn)
	if err := server.DispatchRequest(ctx, conn); !errors.Is(err, wantErr) {
		t.Fatalf("DispatchRequest() error = %v, want invalid frame", err)
	}
	mc, _ := sessionFromContext[any](ctx)
	waitForChannel(t, mc.cleanupDone, "fatal unpack cleanup")
	if got := callbackCount.Load(); got != 1 {
		t.Fatalf("OnCloseFunc calls = %d, want 1", got)
	}
	if !errors.Is(context.Cause(mc.connectionContext()), wantErr) {
		t.Fatalf("connection context cause = %v, want invalid frame", context.Cause(mc.connectionContext()))
	}
	server.globalGate.Close()
	server.cleanupTracker.Close()
}

func TestBaseServerRealNetpollActiveAndPeerClose(t *testing.T) {
	tests := []struct {
		name        string
		activeClose bool
	}{
		{name: "active close after response", activeClose: true},
		{name: "peer close", activeClose: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			closed := make(chan struct{}, 1)
			handled := make(chan struct{}, 1)
			var rawDisconnectCalls atomic.Int32
			server, err := NewBaseServer[any](
				"tcp",
				"127.0.0.1:0",
				WithNetpollOptions[any](
					netpoll.WithOnPrepare(func(netpoll.Connection) context.Context {
						return context.Background()
					}),
					netpoll.WithOnDisconnect(func(context.Context, netpoll.Connection) {
						rawDisconnectCalls.Add(1)
					}),
				),
				WithUnpackFunc[any](func(_ context.Context, reader netpoll.Reader) (protocol.PDU, error) {
					_, err := reader.Next(1)
					return nil, err
				}),
				WithHandleFunc[any](func(context.Context, protocol.PDU) ([]byte, error) {
					handled <- struct{}{}
					if test.activeClose {
						return []byte{0x7f}, errors.New("close after response")
					}
					return nil, nil
				}),
				WithOnCloseFunc[any](func(context.Context, netpoll.Connection) {
					closed <- struct{}{}
				}),
			)
			if err != nil {
				t.Fatalf("NewBaseServer() error = %v", err)
			}
			addr, runResult := startRealServer(t, server)
			client, err := net.Dial("tcp", addr.String())
			if err != nil {
				t.Fatalf("Dial() error = %v", err)
			}
			if _, err := client.Write([]byte{1}); err != nil {
				t.Fatalf("client Write() error = %v", err)
			}
			waitForChannel(t, handled, "real handler")
			if test.activeClose {
				if err := client.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
					t.Fatalf("SetReadDeadline() error = %v", err)
				}
				response := []byte{0}
				if _, err := io.ReadFull(client, response); err != nil {
					t.Fatalf("read response before active close: %v", err)
				}
				if response[0] != 0x7f {
					t.Fatalf("response = %#x, want 0x7f", response[0])
				}
			} else if err := client.Close(); err != nil {
				t.Fatalf("client Close() error = %v", err)
			}
			waitForChannel(t, closed, "real connection cleanup")
			select {
			case <-closed:
				t.Fatal("real connection cleanup ran more than once")
			default:
			}
			if got := rawDisconnectCalls.Load(); got != 0 {
				t.Fatalf("deprecated raw OnDisconnect calls = %d, want 0", got)
			}
			_ = client.Close()
			shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
			if err := server.Shutdown(shutdownCtx); err != nil {
				cancel()
				t.Fatalf("Shutdown() error = %v", err)
			}
			cancel()
			if err := receiveError(t, runResult, "real Run result"); err != nil {
				t.Fatalf("Run() error = %v", err)
			}
		})
	}
}

func TestBaseServerCloseDuringOnRequestCancelsSessionBeforeCallback(t *testing.T) {
	refreshEntered := make(chan context.Context, 1)
	releaseRefresh := make(chan struct{})
	cleanupCalled := make(chan struct{}, 1)
	var refreshOnce sync.Once
	server, err := NewBaseServer[any](
		"tcp",
		"127.0.0.1:0",
		WithRefreshCtxWhenRead[any](func(ctx context.Context) context.Context {
			refreshOnce.Do(func() {
				refreshEntered <- ctx
				<-releaseRefresh
			})
			return ctx
		}),
		WithUnpackFunc[any](func(_ context.Context, reader netpoll.Reader) (protocol.PDU, error) {
			_, err := reader.Next(1)
			return nil, err
		}),
		WithHandleFunc[any](func(context.Context, protocol.PDU) ([]byte, error) {
			return nil, nil
		}),
		WithOnCloseFunc[any](func(context.Context, netpoll.Connection) {
			cleanupCalled <- struct{}{}
		}),
	)
	if err != nil {
		t.Fatalf("NewBaseServer() error = %v", err)
	}
	addr, runResult := startRealServer(t, server)
	client, err := net.Dial("tcp", addr.String())
	if err != nil {
		t.Fatalf("Dial() error = %v", err)
	}
	defer client.Close()
	if _, err := client.Write([]byte{1}); err != nil {
		t.Fatalf("client Write() error = %v", err)
	}

	var requestCtx context.Context
	select {
	case requestCtx = <-refreshEntered:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for blocked OnRequest")
	}
	mc, ok := sessionFromContext[any](requestCtx)
	if !ok {
		t.Fatal("blocked OnRequest context is missing muxConn")
	}
	closeResult := make(chan error, 1)
	go func() { closeResult <- mc.Close() }()
	if err := receiveError(t, closeResult, "Close while OnRequest blocked"); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if !errors.Is(context.Cause(mc.connectionContext()), ErrConnectionClosed) {
		t.Fatalf("connection cause before OnRequest release = %v, want ErrConnectionClosed", context.Cause(mc.connectionContext()))
	}
	if err := mc.Write(context.Background(), []byte("late")); !errors.Is(err, ErrConnectionClosed) {
		t.Fatalf("Write before delayed CloseCallback = %v, want ErrConnectionClosed", err)
	}
	waitForChannel(t, cleanupCalled, "cleanup before OnRequest release")
	close(releaseRefresh)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	if err := server.Shutdown(shutdownCtx); err != nil {
		cancel()
		t.Fatalf("Shutdown() error = %v", err)
	}
	cancel()
	if err := receiveError(t, runResult, "Run after blocked OnRequest close"); err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func startRealServer[T any](t *testing.T, server *BaseServer[T]) (net.Addr, <-chan error) {
	t.Helper()
	bound := make(chan net.Addr, 1)
	server.listenerFactory = func(network, address string) (netpoll.Listener, error) {
		listener, err := netpoll.CreateListener(network, address)
		if err == nil {
			bound <- listener.Addr()
		}
		return listener, err
	}
	runResult := make(chan error, 1)
	go func() { runResult <- server.Run() }()
	select {
	case addr := <-bound:
		return addr, runResult
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for listener bind")
		return nil, nil
	}
}

func receiveError(t *testing.T, result <-chan error, description string) error {
	t.Helper()
	select {
	case err := <-result:
		return err
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for %s", description)
		return nil
	}
}

type fakeEventLoop struct {
	serveStarted    chan struct{}
	shutdownStarted chan struct{}
	serveRelease    chan struct{}
	serveOnce       sync.Once
	shutdownOnce    sync.Once
	serveErr        error
	shutdownErr     error
}

func newFakeEventLoop(serveErr error) *fakeEventLoop {
	return &fakeEventLoop{
		serveStarted:    make(chan struct{}),
		shutdownStarted: make(chan struct{}),
		serveRelease:    make(chan struct{}),
		serveErr:        serveErr,
	}
}

func (e *fakeEventLoop) Serve(net.Listener) error {
	e.serveOnce.Do(func() { close(e.serveStarted) })
	<-e.serveRelease
	return e.serveErr
}

func (e *fakeEventLoop) Shutdown(context.Context) error {
	e.shutdownOnce.Do(func() {
		close(e.shutdownStarted)
		close(e.serveRelease)
	})
	return e.shutdownErr
}

func (e *fakeEventLoop) finishServe() {
	e.shutdownOnce.Do(func() { close(e.serveRelease) })
}

type fakeListener struct {
	addr      net.Addr
	closed    chan struct{}
	closeOnce sync.Once
}

func newFakeListener(address string) *fakeListener {
	return &fakeListener{addr: fakeAddr(address), closed: make(chan struct{})}
}

func (l *fakeListener) Accept() (net.Conn, error) {
	<-l.closed
	return nil, net.ErrClosed
}

func (l *fakeListener) Close() error {
	l.closeOnce.Do(func() { close(l.closed) })
	return nil
}

func (l *fakeListener) Addr() net.Addr { return l.addr }
func (l *fakeListener) Fd() int        { return -1 }
