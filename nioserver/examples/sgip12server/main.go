package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hujm2023/hlog"

	"github.com/hujm2023/go-sms-protocol/nioserver"
	"github.com/hujm2023/go-sms-protocol/nioserver/examples/internal/exampleutil"
	"github.com/hujm2023/go-sms-protocol/sgip"
	"github.com/hujm2023/go-sms-protocol/sgip/sgip12"
)

type config struct {
	exampleutil.ServerConfig
	exampleutil.Credentials
	callbackCredentials exampleutil.Credentials
	callbackAddress     string
	callbackTimeout     time.Duration
	nodeID              uint32
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
	if err := run(ctx, cfg, hlog.DefaultLogger(), nil); err != nil {
		fmt.Fprintf(output, "sgip gateway: %v\n", err)
		return 1
	}
	return 0
}

func parseConfig(args []string, lookup func(string) (string, bool), output io.Writer) (config, error) {
	var cfg config
	var node uint64
	f := flag.NewFlagSet("sgip12server", flag.ContinueOnError)
	f.SetOutput(output)
	cfg.ServerConfig.Flags(f, "127.0.0.1:8801")
	f.StringVar(&cfg.callbackAddress, "sp-address", "127.0.0.1:8802", "configured SP listener for gateway-initiated report connections")
	f.DurationVar(&cfg.callbackTimeout, "callback-timeout", 5*time.Second, "deadline for one callback batch including authentication and acknowledgements")
	f.Uint64Var(&node, "node-id", 2000000001, "gateway node ID in outbound twelve-byte sequences")
	if err := f.Parse(args); err != nil {
		return config{}, err
	}
	if f.NArg() != 0 || node == 0 || node > math.MaxUint32 || cfg.callbackTimeout <= 0 {
		return config{}, errors.New("unexpected arguments, invalid node-id or callback timeout")
	}
	cfg.nodeID = uint32(node)
	credentials, err := exampleutil.LoadCredentials(lookup, "SGIP12", 16, 16)
	if err != nil {
		return config{}, err
	}
	cfg.Credentials = credentials
	cfg.callbackCredentials, err = exampleutil.LoadCredentials(lookup, "SGIP12_CALLBACK", 16, 16)
	if err != nil {
		return config{}, err
	}
	address, err := net.ResolveTCPAddr("tcp", cfg.callbackAddress)
	if err != nil || address == nil || address.Port == 0 {
		return config{}, errors.New("sp-address must resolve to a nonzero TCP port")
	}
	return cfg, cfg.ServerConfig.Validate()
}

func run(ctx context.Context, cfg config, logger hlog.FullLogger, ready func(net.Addr)) error {
	serviceCtx, stop := context.WithCancel(context.Background())
	defer stop()
	h := &serverHandler{cfg: cfg, logger: logger}
	h.receipts = exampleutil.NewQueue(serviceCtx, cfg.Workers, cfg.Queue, h.deliver)
	server, err := nioserver.NewBaseServer[connectionSession](
		"tcp", cfg.Listen,
		nioserver.WithUnpackFunc[connectionSession](exampleutil.Unpack(sgip.HeaderLength, maxFrameSize, sgip12.DecodeSGIP12)),
		nioserver.WithHandleFunc[connectionSession](h.handle),
		nioserver.WithReadTimeout[connectionSession](cfg.ReadTimeout),
		nioserver.WithWriteTimeout[connectionSession](cfg.WriteTimeout),
		nioserver.WithIdleTimeout[connectionSession](cfg.IdleTimeout),
		nioserver.WithCleanupTimeout[connectionSession](cfg.ShutdownTimeout),
		nioserver.WithLogger[connectionSession](logger),
	)
	if err != nil {
		stop()
		waitCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		return errors.Join(err, h.receipts.Wait(waitCtx))
	}
	return exampleutil.Run(ctx, server, logger, exampleutil.Lifecycle{
		Name: "sgip 1.2 example gateway", ShutdownTimeout: cfg.ShutdownTimeout,
		Ready: ready, Stop: stop, Wait: h.receipts.Wait,
	})
}
