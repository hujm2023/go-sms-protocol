package main

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	protocol "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/nioserver"
	"github.com/hujm2023/go-sms-protocol/nioserver/examples/internal/exampleutil"
	"github.com/hujm2023/go-sms-protocol/sgip"
	"github.com/hujm2023/go-sms-protocol/sgip/sgip12"
)

type scriptWire struct {
	io.Reader
	io.Writer
}
type fakeCallbackConn struct {
	nioserver.ISMSConn[callbackSession]
	session callbackSession
	closed  bool
	final   []byte
}

func (c *fakeCallbackConn) GetBizData() callbackSession  { return c.session }
func (c *fakeCallbackConn) SetBizData(s callbackSession) { c.session = s }
func (c *fakeCallbackConn) WriteAndClose(_ context.Context, data []byte) error {
	c.final = append([]byte(nil), data...)
	c.closed = true
	return nil
}

func testEnv(name string) (string, bool) {
	switch name {
	case "SGIP12_ACCOUNT":
		return "sp-demo", true
	case "SGIP12_PASSWORD":
		return "sp-secret", true
	case "SGIP12_CALLBACK_ACCOUNT":
		return "gw-demo", true
	case "SGIP12_CALLBACK_PASSWORD":
		return "gw-secret", true
	default:
		return "", false
	}
}

func TestSubmitConnectionFinishesBeforeWaitingForCallback(t *testing.T) {
	sequence := [3]uint32{3000000001, 1020123456, 2}
	var input, output, log bytes.Buffer
	for _, pdu := range []protocol.PDU{
		&sgip12.BindResp{Header: sgip.Header{CommandID: sgip.SGIP_BIND_REP, Sequence: [3]uint32{sequence[0], sequence[1], 1}}},
		&sgip12.SubmitResp{Header: sgip.Header{CommandID: sgip.SGIP_SUBMIT_REP, Sequence: sequence}},
		&sgip12.UnbindResp{Header: sgip.Header{CommandID: sgip.SGIP_UNBIND_REP, Sequence: [3]uint32{sequence[0], sequence[1], 3}}},
	} {
		if err := exampleutil.Write(&input, pdu); err != nil {
			t.Fatal(err)
		}
	}
	cfg := config{Credentials: exampleutil.Credentials{Account: "sp-demo", Password: "sp-secret"}, nodeID: sequence[0]}
	submit := &sgip12.Submit{Header: sgip.Header{CommandID: sgip.SGIP_SUBMIT, Sequence: sequence}}
	if err := submitSession(scriptWire{Reader: &input, Writer: &output}, cfg, submit, &log); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(log.String(), "awaiting a separate callback connection") {
		t.Fatal("submission flow did not finish independently of callbacks")
	}
}

func TestCallbackAuthenticatesAndEchoesFullReportSequence(t *testing.T) {
	cfg, err := parseConfig(nil, testEnv, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	expected := [3]uint32{3000000001, 1020123456, 2}
	h := &callbackHandler{cfg: cfg, expected: expected, result: make(chan error, 1)}
	conn := &fakeCallbackConn{}
	bind := sgip12.NewBind("gw-demo", "gw-secret", 2000000001, 1)
	bind.Type = sgip.SMG_SP
	if _, err := h.handleConnection(context.Background(), conn, bind); err != nil {
		t.Fatal(err)
	}
	report := &sgip12.Report{Header: sgip.Header{CommandID: sgip.SGIP_REPORT, Sequence: [3]uint32{2000000001, 1020123457, 7}}, SubmitSequence: expected, UserNumber: cfg.destination}
	data, err := h.handleConnection(context.Background(), conn, report)
	if err != nil {
		t.Fatal(err)
	}
	var ack sgip12.ReportResp
	if err := ack.IDecode(data); err != nil {
		t.Fatal(err)
	}
	if ack.Sequence != report.Sequence || ack.Result != 0 {
		t.Fatal("REPORT_RESP lost the report's twelve-byte sequence")
	}
	unbind := &sgip12.Unbind{Header: sgip.Header{CommandID: sgip.SGIP_UNBIND, Sequence: [3]uint32{2000000001, 1020123457, 8}}}
	if _, err := h.handleConnection(context.Background(), conn, unbind); err != nil {
		t.Fatal(err)
	}
	if !conn.closed {
		t.Fatal("callback did not flush and close")
	}
	if err := <-h.result; err != nil {
		t.Fatal(err)
	}
}

func TestCallbackRejectsWrongNodeWithMatchingCounter(t *testing.T) {
	cfg, err := parseConfig(nil, testEnv, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	h := &callbackHandler{cfg: cfg, expected: [3]uint32{3000000001, 1020123456, 2}, result: make(chan error, 1)}
	conn := &fakeCallbackConn{session: callbackSession{bound: true}}
	report := &sgip12.Report{Header: sgip.Header{CommandID: sgip.SGIP_REPORT, Sequence: [3]uint32{2000000001, 1020123457, 7}}, SubmitSequence: [3]uint32{3000000002, 1020123456, 2}, UserNumber: cfg.destination}
	data, err := h.handleConnection(context.Background(), conn, report)
	if err == nil {
		t.Fatal("receipt matched only the last four bytes of the sequence")
	}
	var ack sgip12.ReportResp
	if err := ack.IDecode(data); err != nil {
		t.Fatal(err)
	}
	if ack.Result != sgip.STAT_ERSEQ || conn.session.reportSeen {
		t.Fatal("unrelated report was accepted")
	}
}

func TestCallbackRejectsWrongCredentialsAndDirection(t *testing.T) {
	cfg, err := parseConfig(nil, testEnv, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name     string
		kind     sgip.LoginType
		password string
		want     sgip.RespStatus
	}{
		{name: "password", kind: sgip.SMG_SP, password: "wrong", want: sgip.STAT_ILLLOGIN},
		{name: "direction", kind: sgip.SP_SMG, password: "gw-secret", want: sgip.STAT_ERLGNTYPE},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &callbackHandler{cfg: cfg, result: make(chan error, 1)}
			conn := &fakeCallbackConn{}
			p := sgip12.NewBind("gw-demo", tt.password, 2000000001, 1)
			p.Type = tt.kind
			data, err := h.bind(conn, p)
			if err == nil {
				t.Fatal("invalid callback authenticated")
			}
			var ack sgip12.BindResp
			if err := ack.IDecode(data); err != nil {
				t.Fatal(err)
			}
			if ack.Result != tt.want || conn.session.bound {
				t.Fatal("invalid callback state")
			}
		})
	}
}

func TestConfiguration(t *testing.T) {
	if _, err := parseConfig(nil, testEnv, io.Discard); err != nil {
		t.Fatal(err)
	}
	if _, err := parseConfig([]string{"-destination=13800138000"}, testEnv, io.Discard); err == nil {
		t.Fatal("destination missing country code accepted")
	}
	if realMain([]string{"-help"}, testEnv, io.Discard, io.Discard) != 0 {
		t.Fatal("help failed")
	}
	if realMain([]string{"-timeout=0"}, testEnv, io.Discard, io.Discard) != 2 {
		t.Fatal("configuration exit code")
	}
}
