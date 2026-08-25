package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cloudwego/netpoll"
	"github.com/hujm2023/hlog"

	"github.com/hujm2023/go-sms-protocol/nioserver"
)

const (
	accountEnvironment  = "CMPP20_ACCOUNT"
	passwordEnvironment = "CMPP20_PASSWORD"
	maxGatewayID        = 1<<22 - 1
	startupTimeout      = 5 * time.Second
)

type config struct {
	listenAddress    string
	account          string
	password         string
	gatewayID        uint64
	authSkew         time.Duration
	receiptMinDelay  time.Duration
	receiptMaxDelay  time.Duration
	receiptWorkers   int
	receiptQueueSize int
	heartbeatEvery   time.Duration
	heartbeatMisses  int
	readTimeout      time.Duration
	writeTimeout     time.Duration
	idleTimeout      time.Duration
	shutdownTimeout  time.Duration
}

func main() {
	os.Exit(realMain(os.Args[1:], os.LookupEnv, os.Stderr))
}

func realMain(args []string, lookupEnv func(string) (string, bool), stderr io.Writer) int {
	cfg, err := parseConfig(args, lookupEnv, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err != nil {
		fmt.Fprintf(stderr, "configuration error: %v\n", err)
		return 2
	}

	// 信号处理属于应用层职责。收到 SIGINT/SIGTERM 后统一走有超时的 Shutdown，
	// 而不是让进程直接退出并丢失正在写出的响应。
	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(signalCtx, cfg, hlog.DefaultLogger()); err != nil {
		hlog.Errorf("cmpp 2.0 example server stopped with error: %v", err)
		return 1
	}
	return 0
}

func parseConfig(
	args []string,
	lookupEnv func(string) (string, bool),
	output io.Writer,
) (config, error) {
	var cfg config
	flags := flag.NewFlagSet("cmpp20server", flag.ContinueOnError)
	flags.SetOutput(output)
	flags.StringVar(&cfg.listenAddress, "listen", "127.0.0.1:7890", "TCP address to listen on")
	flags.Uint64Var(&cfg.gatewayID, "gateway-id", 1, "CMPP gateway ID used to generate message IDs (0..4194303)")
	flags.DurationVar(&cfg.authSkew, "auth-skew", 5*time.Minute, "maximum accepted CMPP_CONNECT timestamp skew")
	flags.DurationVar(&cfg.receiptMinDelay, "receipt-min-delay", 100*time.Millisecond, "minimum simulated delivery delay")
	flags.DurationVar(&cfg.receiptMaxDelay, "receipt-max-delay", 500*time.Millisecond, "maximum simulated delivery delay")
	flags.IntVar(&cfg.receiptWorkers, "receipt-workers", 4, "number of delivery-receipt workers")
	flags.IntVar(&cfg.receiptQueueSize, "receipt-queue", 1024, "maximum queued delivery receipts")
	flags.DurationVar(&cfg.heartbeatEvery, "heartbeat-interval", 30*time.Second, "server heartbeat interval")
	flags.IntVar(&cfg.heartbeatMisses, "heartbeat-misses", 3, "unanswered heartbeats before closing a connection")
	flags.DurationVar(&cfg.readTimeout, "read-timeout", 30*time.Second, "per-read transport timeout")
	flags.DurationVar(&cfg.writeTimeout, "write-timeout", 5*time.Second, "per-write transport timeout")
	flags.DurationVar(&cfg.idleTimeout, "idle-timeout", 2*time.Minute, "idle connection timeout")
	flags.DurationVar(&cfg.shutdownTimeout, "shutdown-timeout", 10*time.Second, "graceful shutdown deadline")
	if err := flags.Parse(args); err != nil {
		return config{}, err
	}
	if flags.NArg() != 0 {
		return config{}, fmt.Errorf("unexpected positional arguments: %q", flags.Args())
	}

	// 账号和密码不做命令行参数，避免密码出现在 shell history 和进程参数列表中。
	account, ok := lookupEnv(accountEnvironment)
	if !ok || account == "" {
		return config{}, fmt.Errorf("%s is required", accountEnvironment)
	}
	password, ok := lookupEnv(passwordEnvironment)
	if !ok || password == "" {
		return config{}, fmt.Errorf("%s is required", passwordEnvironment)
	}
	cfg.account = account
	cfg.password = password

	if len(cfg.account) > 6 || strings.ContainsAny(cfg.account, "\x00\r\n\t") {
		return config{}, fmt.Errorf("%s must be 1 to 6 bytes without control characters", accountEnvironment)
	}
	if cfg.listenAddress == "" {
		return config{}, errors.New("listen address must not be empty")
	}
	if _, err := net.ResolveTCPAddr("tcp", cfg.listenAddress); err != nil {
		return config{}, fmt.Errorf("resolve listen address: %w", err)
	}
	if cfg.gatewayID > maxGatewayID {
		return config{}, fmt.Errorf("gateway-id %d exceeds %d", cfg.gatewayID, maxGatewayID)
	}
	if cfg.authSkew <= 0 {
		return config{}, errors.New("auth-skew must be positive")
	}
	if cfg.receiptMinDelay <= 0 || cfg.receiptMaxDelay < cfg.receiptMinDelay {
		return config{}, errors.New("receipt delays must be positive and max must be at least min")
	}
	if cfg.receiptWorkers < 1 {
		return config{}, errors.New("receipt-workers must be positive")
	}
	if cfg.receiptQueueSize < 99 {
		// CMPP 2.0 单次 SUBMIT 最多包含 99 个号码。队列至少容纳一整批，
		// 才能实现“整批入队或整批拒绝”的处理语义。
		return config{}, errors.New("receipt-queue must hold at least one 99-recipient submit")
	}
	if cfg.heartbeatEvery <= 0 || cfg.heartbeatMisses < 1 {
		return config{}, errors.New("heartbeat interval and misses must be positive")
	}
	if cfg.readTimeout <= 0 || cfg.writeTimeout <= 0 || cfg.idleTimeout <= 0 {
		return config{}, errors.New("transport timeouts must be positive")
	}
	if cfg.shutdownTimeout <= 0 {
		return config{}, errors.New("shutdown-timeout must be positive")
	}
	return cfg, nil
}

func run(ctx context.Context, cfg config, logger hlog.FullLogger) error {
	// 后台回执和服务端探活必须使用独立的服务级 context。
	// nioserver 的请求 context 会随单次处理或连接关闭而取消，不能承载异步任务。
	serviceCtx, stopServices := context.WithCancel(context.Background())
	receipts := newReceiptDispatcher(
		serviceCtx,
		logger,
		cfg.receiptWorkers,
		cfg.receiptQueueSize,
		cfg.receiptMinDelay,
		cfg.receiptMaxDelay,
		cfg.writeTimeout,
	)
	heartbeats := newHeartbeatManager(
		serviceCtx,
		logger,
		cfg.heartbeatEvery,
		cfg.heartbeatMisses,
		cfg.writeTimeout,
	)
	handler := &serverHandler{
		account:    cfg.account,
		password:   cfg.password,
		authSkew:   cfg.authSkew,
		logger:     logger,
		receipts:   receipts,
		heartbeats: heartbeats,
		messageIDs: messageIDGenerator{gatewayID: cfg.gatewayID},
		now:        time.Now,
	}

	server, err := nioserver.NewBaseServer[connectionSession](
		"tcp",
		cfg.listenAddress,
		nioserver.WithUnpackFunc[connectionSession](newUnpackFunc(maxCMPP20FrameSize)),
		nioserver.WithHandleFunc[connectionSession](handler.handle),
		nioserver.WithReadTimeout[connectionSession](cfg.readTimeout),
		nioserver.WithWriteTimeout[connectionSession](cfg.writeTimeout),
		nioserver.WithIdleTimeout[connectionSession](cfg.idleTimeout),
		nioserver.WithCleanupTimeout[connectionSession](cfg.shutdownTimeout),
		nioserver.WithLogger[connectionSession](logger),
		nioserver.WithOnCloseFunc[connectionSession](func(closeCtx context.Context, _ netpoll.Connection) {
			// OnClose 只清理连接索引；回执 worker 自己处理写入已关闭连接时的错误。
			// 不依赖 netpoll CloseCallback 执行业务收尾，避免关闭回调时序影响后台生命周期。
			if conn, ok := nioserver.GetCtxConn[connectionSession](closeCtx); ok {
				heartbeats.unregister(conn)
			}
		}),
	)
	if err != nil {
		stopServices()
		waitCtx, cancel := context.WithTimeout(context.Background(), cfg.shutdownTimeout)
		defer cancel()
		return errors.Join(fmt.Errorf("create cmpp server: %w", err), waitForServices(waitCtx, receipts, heartbeats))
	}

	// 使用容量为 1 的 channel，确保调用方因启动超时提前返回时，Run goroutine 也不会卡在上报结果。
	runResult := make(chan error, 1)
	go func() {
		runResult <- server.Run()
	}()

	// BaseServer.Run 没有单独的 ready channel，因此在有上限的时间内观察 Addr。
	// 同时监听 Run 的结果，绑定失败时可以立即返回真实错误。
	startupTimer := time.NewTimer(startupTimeout)
	startupTicker := time.NewTicker(10 * time.Millisecond)
	for server.Addr() == nil {
		select {
		case err := <-runResult:
			startupTimer.Stop()
			startupTicker.Stop()
			stopServices()
			waitCtx, cancel := context.WithTimeout(context.Background(), cfg.shutdownTimeout)
			defer cancel()
			if err == nil {
				err = errors.New("cmpp server stopped before binding")
			}
			return errors.Join(err, waitForServices(waitCtx, receipts, heartbeats))
		case <-ctx.Done():
			startupTimer.Stop()
			startupTicker.Stop()
			return shutdown(server, runResult, stopServices, receipts, heartbeats, cfg.shutdownTimeout)
		case <-startupTimer.C:
			startupTicker.Stop()
			shutdownErr := shutdown(server, runResult, stopServices, receipts, heartbeats, cfg.shutdownTimeout)
			return errors.Join(errors.New("timed out waiting for cmpp server to bind"), shutdownErr)
		case <-startupTicker.C:
		}
	}
	startupTimer.Stop()
	startupTicker.Stop()
	logger.Infof("cmpp 2.0 example server listening address=%s", server.Addr())

	select {
	case err := <-runResult:
		stopServices()
		waitCtx, cancel := context.WithTimeout(context.Background(), cfg.shutdownTimeout)
		defer cancel()
		if err == nil {
			err = errors.New("cmpp server stopped unexpectedly")
		}
		return errors.Join(err, waitForServices(waitCtx, receipts, heartbeats))
	case <-ctx.Done():
		logger.Infof("cmpp 2.0 example server shutting down")
		return shutdown(server, runResult, stopServices, receipts, heartbeats, cfg.shutdownTimeout)
	}
}

func shutdown(
	server *nioserver.BaseServer[connectionSession],
	runResult <-chan error,
	stopServices context.CancelFunc,
	receipts *receiptDispatcher,
	heartbeats *heartbeatManager,
	timeout time.Duration,
) error {
	// 先取消后台任务，阻止新探活和新回执；再关闭监听及连接；最后等待所有 goroutine 退出。
	// 三部分共享同一个截止时间，避免任一阶段让进程无限阻塞。
	stopServices()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	serverErr := server.Shutdown(shutdownCtx)
	servicesErr := waitForServices(shutdownCtx, receipts, heartbeats)
	var runErr error
	select {
	case runErr = <-runResult:
	case <-shutdownCtx.Done():
		runErr = shutdownCtx.Err()
	}
	return errors.Join(serverErr, servicesErr, runErr)
}

func waitForServices(
	ctx context.Context,
	receipts *receiptDispatcher,
	heartbeats *heartbeatManager,
) error {
	return errors.Join(receipts.wait(ctx), heartbeats.wait(ctx))
}
