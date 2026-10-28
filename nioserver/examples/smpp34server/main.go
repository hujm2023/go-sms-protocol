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
	"syscall"
	"time"

	"github.com/cloudwego/netpoll"

	smslog "github.com/hujm2023/go-sms-protocol/logger"
	"github.com/hujm2023/go-sms-protocol/nioserver"
	"github.com/hujm2023/go-sms-protocol/nioserver/examples/internal/exampleutil"
	"github.com/hujm2023/go-sms-protocol/smpp"
	"github.com/hujm2023/go-sms-protocol/smpp/smpp34"
)

type config struct {
	exampleutil.ServerConfig
	exampleutil.Credentials
	heartbeatEvery  time.Duration
	heartbeatMisses int
}

func main() { os.Exit(realMain(os.Args[1:], os.LookupEnv, os.Stderr)) }

func realMain(args []string, lookup func(string) (string, bool), output io.Writer) int {
	cfg, err := parseConfig(args, lookup, output)
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err != nil {
		fmt.Fprintf(output, "configuration error: %v\n", err)
		return 2
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, cfg, smslog.DefaultLogger(), nil); err != nil {
		fmt.Fprintf(output, "smpp server: %v\n", err)
		return 1
	}
	return 0
}

func parseConfig(args []string, lookup func(string) (string, bool), output io.Writer) (config, error) {
	var cfg config
	f := flag.NewFlagSet("smpp34server", flag.ContinueOnError)
	f.SetOutput(output)
	cfg.ServerConfig.Flags(f, "127.0.0.1:2775")
	f.DurationVar(&cfg.heartbeatEvery, "heartbeat-interval", 30*time.Second, "server enquire_link interval")
	f.IntVar(&cfg.heartbeatMisses, "heartbeat-misses", 3, "unanswered enquire_link requests before closing")
	if err := f.Parse(args); err != nil {
		return config{}, err
	}
	if f.NArg() != 0 {
		return config{}, errors.New("unexpected positional arguments")
	}
	credentials, err := exampleutil.LoadCredentials(lookup, "SMPP34", 15, 8)
	if err != nil {
		return config{}, err
	}
	cfg.Credentials = credentials
	if err := smpp.ValidateCString(smpp.SYSTEM_ID, cfg.Account, 16); err != nil {
		return config{}, err
	}
	if err := smpp.ValidateCString(smpp.PASSWORD, cfg.Password, 9); err != nil {
		return config{}, err
	}
	if cfg.heartbeatEvery <= 0 || cfg.heartbeatMisses < 1 {
		return config{}, errors.New("heartbeat interval and attempts must be positive")
	}
	return cfg, cfg.ServerConfig.Validate()
}

func run(ctx context.Context, cfg config, logger smslog.FullLogger, ready func(net.Addr)) error {
	serviceCtx, stop := context.WithCancel(context.Background())
	defer stop()
	heartbeats := newHeartbeatManager(serviceCtx, logger, cfg.heartbeatEvery, cfg.heartbeatMisses, cfg.WriteTimeout)
	h := &serverHandler{cfg: cfg, logger: logger, heartbeats: heartbeats}
	h.receipts = exampleutil.NewQueue(serviceCtx, cfg.Workers, cfg.Queue, h.deliver)
	wait := func(ctx context.Context) error { return errors.Join(h.receipts.Wait(ctx), heartbeats.wait(ctx)) }
	server, err := nioserver.NewBaseServer[connectionSession](
		"tcp", cfg.Listen,
		nioserver.WithUnpackFunc[connectionSession](exampleutil.Unpack(16, smpp.MAX_PDU_SIZE, smpp34.DecodeSMPP34)),
		nioserver.WithHandleFunc[connectionSession](h.handle),
		nioserver.WithReadTimeout[connectionSession](cfg.ReadTimeout),
		nioserver.WithWriteTimeout[connectionSession](cfg.WriteTimeout),
		nioserver.WithIdleTimeout[connectionSession](cfg.IdleTimeout),
		nioserver.WithCleanupTimeout[connectionSession](cfg.ShutdownTimeout),
		nioserver.WithLogger[connectionSession](logger),
		nioserver.WithOnCloseFunc[connectionSession](func(ctx context.Context, _ netpoll.Connection) {
			if conn, ok := nioserver.GetCtxConn[connectionSession](ctx); ok {
				heartbeats.unregister(conn)
			}
		}),
	)
	if err != nil {
		stop()
		waitCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		return errors.Join(err, wait(waitCtx))
	}
	return exampleutil.Run(ctx, server, logger, exampleutil.Lifecycle{
		Name: "smpp 3.4 example server", ShutdownTimeout: cfg.ShutdownTimeout,
		Ready: ready, Stop: stop, Wait: wait,
	})
}
