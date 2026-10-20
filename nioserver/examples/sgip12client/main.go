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

	protocol "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/nioserver"
	"github.com/hujm2023/go-sms-protocol/nioserver/examples/internal/exampleutil"
	"github.com/hujm2023/go-sms-protocol/sgip"
	"github.com/hujm2023/go-sms-protocol/sgip/sgip12"
)

const maxFrameSize = 64 * 1024

type config struct {
	exampleutil.Credentials
	callbackCredentials                           exampleutil.Credentials
	address, listen, source, destination, message string
	nodeID                                        uint32
	timeout                                       time.Duration
}

func main() { os.Exit(realMain(os.Args[1:], os.LookupEnv, os.Stdout, os.Stderr)) }

func realMain(args []string, lookup func(string) (string, bool), stdout, stderr io.Writer) int {
	cfg, err := parseConfig(args, lookup, stderr)
	if errors.Is(err, flag.ErrHelp) {
		return 0
	}
	if err != nil {
		fmt.Fprintf(stderr, "configuration error: %v\n", err)
		return 2
	}
	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ctx, cancel := context.WithTimeout(signalCtx, cfg.timeout)
	defer cancel()
	if err := runClient(ctx, cfg, stdout, nil); err != nil {
		fmt.Fprintf(stderr, "sgip SP: %v\n", err)
		return 1
	}
	return 0
}

func parseConfig(args []string, lookup func(string) (string, bool), output io.Writer) (config, error) {
	var cfg config
	var node uint64
	f := flag.NewFlagSet("sgip12client", flag.ContinueOnError)
	f.SetOutput(output)
	f.StringVar(&cfg.address, "server", "127.0.0.1:8801", "SGIP gateway address")
	f.StringVar(&cfg.listen, "listen", "127.0.0.1:8802", "SP listener for gateway-initiated reports; must match the gateway's -sp-address")
	f.StringVar(&cfg.source, "source", "106900001", "SP access number")
	f.StringVar(&cfg.destination, "destination", "8613800138000", "recipient including the 86 country code")
	f.StringVar(&cfg.message, "message", "hello SGIP", "ASCII message, at most 140 bytes")
	f.Uint64Var(&node, "node-id", 3000000001, "SP node ID in twelve-byte command sequences")
	f.DurationVar(&cfg.timeout, "timeout", 30*time.Second, "deadline for submission and callback completion")
	if err := f.Parse(args); err != nil {
		return config{}, err
	}
	if f.NArg() != 0 || cfg.timeout <= 0 || node == 0 || node > math.MaxUint32 {
		return config{}, errors.New("unexpected arguments, invalid node-id or timeout")
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
	for _, value := range []string{cfg.source, cfg.destination} {
		if err := exampleutil.TextField(value, 1, 21); err != nil {
			return config{}, err
		}
	}
	if len(cfg.destination) <= 2 || cfg.destination[:2] != "86" {
		return config{}, errors.New("SGIP destination must include country code 86")
	}
	return cfg, exampleutil.ASCIIMessage(cfg.message, 140)
}

func runClient(ctx context.Context, cfg config, output io.Writer, ready func(net.Addr)) (returnErr error) {
	submit := &sgip12.Submit{
		Header:   sgip.NewHeader(0, sgip.SGIP_SUBMIT, cfg.nodeID, 2),
		SpNumber: cfg.source, ChargeNumber: "000000000000000000000",
		UserCount: 1, UserNumber: []string{cfg.destination}, CorpID: "00001",
		ReportFlag: 1, MessageLength: uint32(len(cfg.message)), MessageContent: []byte(cfg.message),
	}
	handler := &callbackHandler{cfg: cfg, expected: submit.Sequence, result: make(chan error, 1)}
	server, err := nioserver.NewBaseServer[callbackSession](
		"tcp", cfg.listen,
		nioserver.WithUnpackFunc[callbackSession](exampleutil.Unpack(sgip.HeaderLength, maxFrameSize, sgip12.DecodeSGIP12)),
		nioserver.WithHandleFunc[callbackSession](handler.handle),
		nioserver.WithReadTimeout[callbackSession](10*time.Second),
		nioserver.WithWriteTimeout[callbackSession](5*time.Second),
		nioserver.WithIdleTimeout[callbackSession](10*time.Second),
	)
	if err != nil {
		return err
	}
	callbackCtx, stop := context.WithCancel(context.Background())
	result := make(chan error, 1)
	bound := make(chan net.Addr, 1)
	go func() {
		result <- exampleutil.Run(callbackCtx, server, hlog.DefaultLogger(), exampleutil.Lifecycle{
			Name: "sgip SP callback listener", ShutdownTimeout: 5 * time.Second,
			Ready: func(address net.Addr) { bound <- address },
		})
	}()
	var consumed bool
	defer func() {
		stop()
		if !consumed {
			returnErr = errors.Join(returnErr, <-result)
		}
	}()
	select {
	case address := <-bound:
		fmt.Fprintf(output, "SP callback listener ready address=%s\n", address)
		if ready != nil {
			ready(address)
		}
	case err := <-result:
		consumed = true
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
	if err := submitMessage(ctx, cfg, submit, output); err != nil {
		return err
	}
	select {
	case err := <-handler.result:
		if err != nil {
			return err
		}
		fmt.Fprintf(output, "callback bind accepted type=2\nreport received submit_id=%s state=0 acknowledged\ncallback unbind acknowledged\n",
			sgip.SequenceIDString(submit.Sequence))
		return nil
	case err := <-result:
		consumed = true
		return fmt.Errorf("callback listener stopped before report completion: %w", err)
	case <-ctx.Done():
		return ctx.Err()
	}
}

func submitMessage(ctx context.Context, cfg config, submit *sgip12.Submit, output io.Writer) error {
	conn, cleanup, err := exampleutil.Dial(ctx, cfg.address)
	if err != nil {
		return err
	}
	defer cleanup()
	return submitSession(conn, cfg, submit, output)
}

func submitSession(conn io.ReadWriter, cfg config, submit *sgip12.Submit, output io.Writer) error {
	exchange := func(pdu protocol.PDU) (protocol.PDU, error) {
		if err := exampleutil.Write(conn, pdu); err != nil {
			return nil, err
		}
		return exampleutil.Read(conn, sgip.HeaderLength, maxFrameSize, sgip12.DecodeSGIP12)
	}
	bind := sgip12.NewBind(cfg.Account, cfg.Password, cfg.nodeID, 1)
	bind.Sequence = submit.Sequence
	bind.Sequence[2] = 1
	pdu, err := exchange(bind)
	if err != nil {
		return err
	}
	bindResponse, ok := pdu.(*sgip12.BindResp)
	if !ok || bindResponse.Sequence != bind.Sequence || bindResponse.Result != sgip.STAT_OK {
		return errors.New("submission bind rejected or response sequence mismatch")
	}
	fmt.Fprintln(output, "submission bind accepted type=1")
	pdu, err = exchange(submit)
	if err != nil {
		return err
	}
	response, ok := pdu.(*sgip12.SubmitResp)
	if !ok || response.Sequence != submit.Sequence || response.Result != sgip.STAT_OK {
		return errors.New("submit rejected or response sequence mismatch")
	}
	fmt.Fprintf(output, "submit accepted submit_id=%s\n", sgip.SequenceIDString(submit.Sequence))
	unbind := &sgip12.Unbind{Header: sgip.NewHeader(0, sgip.SGIP_UNBIND, cfg.nodeID, 3)}
	unbind.Sequence = submit.Sequence
	unbind.Sequence[2] = 3
	pdu, err = exchange(unbind)
	if err != nil {
		return err
	}
	unbound, ok := pdu.(*sgip12.UnbindResp)
	if !ok || unbound.Sequence != unbind.Sequence {
		return errors.New("submission unbind response mismatch")
	}
	fmt.Fprintln(output, "submission unbind acknowledged; awaiting a separate callback connection")
	return nil
}
