//go:build integration

package main

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	protocol "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/nioserver/examples/internal/exampleutil"
	"github.com/hujm2023/go-sms-protocol/sgip"
	"github.com/hujm2023/go-sms-protocol/sgip/sgip12"
)

func emulateGateway(ctx context.Context, listener net.Listener, callback <-chan net.Addr, reject bool) error {
	conn, err := listener.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
		return err
	}
	var submitSequence [3]uint32
	for i := 0; i < 3; i++ {
		pdu, err := exampleutil.Read(conn, sgip.HeaderLength, maxFrameSize, sgip12.DecodeSGIP12)
		if err != nil {
			return err
		}
		switch p := pdu.(type) {
		case *sgip12.Bind:
			if p.Type != sgip.SP_SMG {
				return errors.New("wrong submission bind type")
			}
		case *sgip12.Submit:
			submitSequence = p.Sequence
		case *sgip12.Unbind:
		default:
			return errors.New("unexpected submission command")
		}
		if err := exampleutil.Write(conn, pdu.GenEmptyResponse()); err != nil {
			return err
		}
	}
	var address net.Addr
	select {
	case address = <-callback:
	case <-ctx.Done():
		return ctx.Err()
	}
	reportConn, cleanup, err := exampleutil.Dial(ctx, address.String())
	if err != nil {
		return err
	}
	defer cleanup()
	password := "gw-secret"
	if reject {
		password = "wrong"
	}
	bind := sgip12.NewBind("gw-demo", password, 2000000001, 1)
	bind.Type = sgip.SMG_SP
	if err := exampleutil.Write(reportConn, bind); err != nil {
		return err
	}
	pdu, err := exampleutil.Read(reportConn, sgip.HeaderLength, maxFrameSize, sgip12.DecodeSGIP12)
	if err != nil {
		return err
	}
	bound, ok := pdu.(*sgip12.BindResp)
	if !ok || bound.Sequence != bind.Sequence {
		return errors.New("bad callback bind response")
	}
	if reject {
		if bound.Result != sgip.STAT_ILLLOGIN {
			return errors.New("bad callback password accepted")
		}
		return nil
	}
	if bound.Result != 0 {
		return errors.New("callback unexpectedly rejected")
	}
	for _, request := range []protocol.PDU{
		&sgip12.Report{Header: sgip.NewHeader(0, sgip.SGIP_REPORT, 2000000001, 2), SubmitSequence: submitSequence, UserNumber: "8613800138000"},
		&sgip12.Unbind{Header: sgip.NewHeader(0, sgip.SGIP_UNBIND, 2000000001, 3)},
	} {
		if err := exampleutil.Write(reportConn, request); err != nil {
			return err
		}
		response, err := exampleutil.Read(reportConn, sgip.HeaderLength, maxFrameSize, sgip12.DecodeSGIP12)
		if err != nil {
			return err
		}
		switch p := response.(type) {
		case *sgip12.ReportResp:
			if p.Result != 0 {
				return errors.New("report rejected")
			}
		case *sgip12.UnbindResp:
		default:
			return errors.New("unexpected callback response")
		}
	}
	return nil
}

func TestTCPClientOwnsCallbackListenerAndAuthenticatesGateway(t *testing.T) {
	tests := []struct {
		name   string
		reject bool
	}{
		{name: "complete exchange"},
		{name: "reject incorrect callback password", reject: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			defer listener.Close()
			cfg, err := parseConfig([]string{"-server=" + listener.Addr().String(), "-listen=127.0.0.1:0"}, testEnv, io.Discard)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			callback := make(chan net.Addr, 1)
			gatewayResult := make(chan error, 1)
			go func() { gatewayResult <- emulateGateway(ctx, listener, callback, tt.reject) }()
			var output strings.Builder
			err = runClient(ctx, cfg, &output, func(address net.Addr) { callback <- address })
			if (err != nil) != tt.reject {
				t.Fatalf("client error=%v reject=%t", err, tt.reject)
			}
			select {
			case err := <-gatewayResult:
				if err != nil {
					t.Fatal(err)
				}
			case <-ctx.Done():
				t.Fatal("gateway goroutine did not stop")
			}
			if !tt.reject && !strings.Contains(output.String(), "callback unbind acknowledged") {
				t.Fatal("callback exchange incomplete")
			}
		})
	}
}
