//go:build integration

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	protocol "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/nioserver/examples/internal/exampleutil"
	"github.com/hujm2023/go-sms-protocol/sgip"
	"github.com/hujm2023/go-sms-protocol/sgip/sgip12"
	"github.com/hujm2023/hlog"
)

func startNetworkServer(t *testing.T, callback string) string {
	t.Helper()
	cfg, err := parseConfig([]string{"-listen=127.0.0.1:0", "-sp-address=" + callback, "-receipt-min-delay=1ms", "-receipt-max-delay=1ms"}, testEnv, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	ready := make(chan net.Addr, 1)
	result := make(chan error, 1)
	go func() { result <- run(ctx, cfg, hlog.DefaultLogger(), func(address net.Addr) { ready <- address }) }()
	t.Cleanup(func() {
		cancel()
		select {
		case err := <-result:
			if err != nil {
				t.Error(err)
			}
		case <-time.After(2 * time.Second):
			t.Error("gateway or callback workers did not stop")
		}
	})
	select {
	case address := <-ready:
		return address.String()
	case <-time.After(time.Second):
		t.Fatal("gateway did not bind")
	}
	return ""
}

func tcpExchange(t *testing.T, conn net.Conn, pdu protocol.PDU) protocol.PDU {
	t.Helper()
	if err := exampleutil.Write(conn, pdu); err != nil {
		t.Fatal(err)
	}
	response, err := exampleutil.Read(conn, sgip.HeaderLength, maxFrameSize, sgip12.DecodeSGIP12)
	if err != nil {
		t.Fatal(err)
	}
	return response
}

func receiveCallback(conn net.Conn, expected [3]uint32) error {
	if err := conn.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
		return err
	}
	read := func() (protocol.PDU, error) {
		return exampleutil.Read(conn, sgip.HeaderLength, maxFrameSize, sgip12.DecodeSGIP12)
	}
	pdu, err := read()
	if err != nil {
		return err
	}
	bind, ok := pdu.(*sgip12.Bind)
	if !ok || bind.Type != sgip.SMG_SP || bind.Name != "gw-demo" || bind.Password != "gw-secret" {
		return errors.New("callback must authenticate with type 2 and gateway credentials")
	}
	if err := exampleutil.Write(conn, bind.GenEmptyResponse()); err != nil {
		return err
	}
	for i := 0; i < 2; i++ {
		pdu, err = read()
		if err != nil {
			return err
		}
		report, ok := pdu.(*sgip12.Report)
		if !ok || report.SubmitSequence != expected || report.State != 0 || report.ReportType != 0 || report.Sequence[0] != 2000000001 {
			return fmt.Errorf("wrong callback report: %T", pdu)
		}
		if err := exampleutil.Write(conn, report.GenEmptyResponse()); err != nil {
			return err
		}
	}
	pdu, err = read()
	if err != nil {
		return err
	}
	unbind, ok := pdu.(*sgip12.Unbind)
	if !ok {
		return errors.New("missing callback UNBIND")
	}
	return exampleutil.Write(conn, unbind.GenEmptyResponse())
}

func TestTCPReportsUseIndependentAuthenticatedConnection(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	request := validSubmit()
	gate := make(chan struct{})
	var once sync.Once
	release := func() { once.Do(func() { close(gate) }) }
	defer release()
	callbackResult := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			callbackResult <- err
			return
		}
		defer conn.Close()
		<-gate // Read the callback only after the submission connection has closed.
		callbackResult <- receiveCallback(conn, request.Sequence)
	}()
	address := startNetworkServer(t, listener.Addr().String())
	conn, err := net.DialTimeout("tcp", address, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatal(err)
	}
	bind := sgip12.NewBind("sp-demo", "sp-secret", 3000000001, 1)
	if p, ok := tcpExchange(t, conn, bind).(*sgip12.BindResp); !ok || p.Result != 0 {
		t.Fatal("submission bind failed")
	}
	if p, ok := tcpExchange(t, conn, request).(*sgip12.SubmitResp); !ok || p.Result != 0 || p.Sequence != request.Sequence {
		t.Fatal("submit response mismatch")
	}
	unbind := &sgip12.Unbind{Header: sgip.NewHeader(0, sgip.SGIP_UNBIND, 3000000001, 3)}
	if p, ok := tcpExchange(t, conn, unbind).(*sgip12.UnbindResp); !ok || p.Sequence != unbind.Sequence {
		t.Fatal("submission unbind mismatch")
	}
	var b [1]byte
	if _, err := conn.Read(b[:]); err != io.EOF {
		t.Fatalf("submission connection did not close: %v", err)
	}
	release()
	select {
	case err := <-callbackResult:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("gateway did not complete the independent report connection")
	}
}
