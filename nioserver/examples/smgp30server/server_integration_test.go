//go:build integration

package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	protocol "github.com/hujm2023/go-sms-protocol"
	smslog "github.com/hujm2023/go-sms-protocol/logger"
	"github.com/hujm2023/go-sms-protocol/smgp"
	"github.com/hujm2023/go-sms-protocol/smgp/smgp30"
)

type readyLogger struct {
	smslog.FullLogger
	ready chan string
}

func (l readyLogger) Infof(format string, args ...interface{}) {
	if strings.HasPrefix(format, "smgp 3.0 example server listening") {
		l.ready <- fmt.Sprint(args[0])
	}
	l.FullLogger.Infof(format, args...)
}

func startExampleServer(t *testing.T) string {
	t.Helper()
	cfg, err := parseConfig([]string{
		"-listen=127.0.0.1:0", "-heartbeat-interval=20ms", "-heartbeat-misses=3",
		"-receipt-min-delay=1ms", "-receipt-max-delay=1ms",
	}, testEnvironment, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	logger := readyLogger{FullLogger: smslog.DefaultLogger(), ready: make(chan string, 1)}
	go func() { result <- run(ctx, cfg, logger) }()
	t.Cleanup(func() {
		cancel()
		select {
		case err := <-result:
			if err != nil {
				t.Errorf("shutdown: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Error("server or background workers did not stop")
		}
	})
	select {
	case address := <-logger.ready:
		return address
	case err := <-result:
		t.Fatalf("server failed to start: %v", err)
	case <-time.After(time.Second):
		t.Fatal("server did not become ready")
	}
	return ""
}

func dialExample(t *testing.T, address string) net.Conn {
	t.Helper()
	conn, err := net.DialTimeout("tcp", address, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	if err := conn.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatal(err)
	}
	return conn
}

func sendExamplePDU(t *testing.T, conn net.Conn, pdu protocol.PDU) {
	t.Helper()
	data, err := pdu.IEncode()
	if err != nil {
		t.Fatal(err)
	}
	// Split the length prefix and body across writes to exercise real framing.
	for _, part := range [][]byte{data[:2], data[2:]} {
		if _, err := io.Copy(conn, strings.NewReader(string(part))); err != nil {
			t.Fatal(err)
		}
	}
}

func receiveExamplePDU(t *testing.T, conn net.Conn) protocol.PDU {
	t.Helper()
	var prefix [4]byte
	if _, err := io.ReadFull(conn, prefix[:]); err != nil {
		t.Fatal(err)
	}
	size := binary.BigEndian.Uint32(prefix[:])
	if size < 12 || size > maxSMGP30FrameSize {
		t.Fatalf("invalid response length: %d", size)
	}
	frame := make([]byte, int(size))
	copy(frame, prefix[:])
	if _, err := io.ReadFull(conn, frame[4:]); err != nil {
		t.Fatal(err)
	}
	pdu, err := smgp30.DecodeSMGP30(frame)
	if err != nil {
		t.Fatal(err)
	}
	return pdu
}

func TestServerTCPSubmitReceiptHeartbeatAndExit(t *testing.T) {
	address := startExampleServer(t)
	conn := dialExample(t, address)
	sendExamplePDU(t, conn, smgp30.NewLogin("900001", "test-secret", 1))
	login, ok := receiveExamplePDU(t, conn).(*smgp30.LoginResp)
	if !ok || login.Status != 0 {
		t.Fatalf("login response = %+v", login)
	}
	sendExamplePDU(t, conn, validSubmit())
	var submit *smgp30.SubmitResp
	var report *smgp30.Deliver
	var serverHeartbeat bool
	for submit == nil || report == nil || !serverHeartbeat {
		switch p := receiveExamplePDU(t, conn).(type) {
		case *smgp30.SubmitResp:
			submit = p
		case *smgp30.Deliver:
			report = p
			ack := p.GenEmptyResponse().(*smgp30.DeliverResp)
			ack.MsgID = p.MsgID
			sendExamplePDU(t, conn, ack)
		case *smgp30.ActiveTest:
			serverHeartbeat = true
			sendExamplePDU(t, conn, p.GenEmptyResponse())
		default:
			t.Fatalf("unexpected response %T", p)
		}
	}
	var receipt smgp30.DeliveryReceipt
	if err := receipt.IDecode(report.MsgContent); err != nil {
		t.Fatal(err)
	}
	if submit.Status != 0 || receipt.ID != submit.MsgID || receipt.Stat != "DELIVRD" {
		t.Fatalf("submit/report mismatch: %+v / %+v", submit, receipt)
	}
	sendExamplePDU(t, conn, &smgp30.ActiveTest{Header: smgp.NewHeader(0, smgp.CommandActiveTest, 3)})
	for {
		pdu := receiveExamplePDU(t, conn)
		if p, ok := pdu.(*smgp30.ActiveTest); ok {
			sendExamplePDU(t, conn, p.GenEmptyResponse())
			continue
		}
		response, ok := pdu.(*smgp30.ActiveTestResp)
		if !ok || response.GetSequenceID() != 3 || response.TotalLength != 12 {
			t.Fatalf("heartbeat must be a header-only response: %+v", pdu)
		}
		break
	}
	sendExamplePDU(t, conn, &smgp30.Exit{Header: smgp.NewHeader(0, smgp.CommandExit, 4)})
	for {
		pdu := receiveExamplePDU(t, conn)
		if p, ok := pdu.(*smgp30.ActiveTest); ok {
			sendExamplePDU(t, conn, p.GenEmptyResponse())
			continue
		}
		if p, ok := pdu.(*smgp30.ExitResp); !ok || p.GetSequenceID() != 4 {
			t.Fatalf("exit response = %+v", pdu)
		}
		break
	}
	var b [1]byte
	if _, err := conn.Read(b[:]); err != io.EOF {
		t.Fatalf("connection not closed after EXIT_RESP: %v", err)
	}
}

func TestServerTCPRejectsInvalidSessions(t *testing.T) {
	address := startExampleServer(t)
	t.Run("bad password gets error before close", func(t *testing.T) {
		conn := dialExample(t, address)
		sendExamplePDU(t, conn, smgp30.NewLogin("900001", "wrong-secret", 1))
		pdu := receiveExamplePDU(t, conn)
		response, ok := pdu.(*smgp30.LoginResp)
		if !ok || response.Status != 21 {
			t.Fatalf("bad-password response = %+v", pdu)
		}
		var b [1]byte
		if _, err := conn.Read(b[:]); err != io.EOF {
			t.Fatalf("rejected login connection not closed: %v", err)
		}
	})
	t.Run("submit before login is closed", func(t *testing.T) {
		conn := dialExample(t, address)
		sendExamplePDU(t, conn, validSubmit())
		var b [1]byte
		if _, err := conn.Read(b[:]); err != io.EOF {
			t.Fatalf("unauthenticated connection not closed: %v", err)
		}
	})
	t.Run("unanswered heartbeat closes connection", func(t *testing.T) {
		conn := dialExample(t, address)
		sendExamplePDU(t, conn, smgp30.NewLogin("900001", "test-secret", 1))
		if p, ok := receiveExamplePDU(t, conn).(*smgp30.LoginResp); !ok || p.Status != 0 {
			t.Fatal("login failed")
		}
		for i := 0; i < 3; i++ {
			if _, ok := receiveExamplePDU(t, conn).(*smgp30.ActiveTest); !ok {
				t.Fatal("expected heartbeat")
			}
		}
		var b [1]byte
		if _, err := conn.Read(b[:]); err != io.EOF {
			t.Fatalf("unresponsive connection not closed: %v", err)
		}
	})
}
