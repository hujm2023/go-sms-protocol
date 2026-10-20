package main

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	protocol "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/nioserver"
	"github.com/hujm2023/go-sms-protocol/nioserver/examples/internal/exampleutil"
	"github.com/hujm2023/go-sms-protocol/sgip"
	"github.com/hujm2023/go-sms-protocol/sgip/sgip12"
	"github.com/hujm2023/hlog"
)

type fakeConn struct {
	nioserver.ISMSConn[connectionSession]
	session connectionSession
}

func (c *fakeConn) GetBizData() connectionSession  { return c.session }
func (c *fakeConn) SetBizData(s connectionSession) { c.session = s }
func (c *fakeConn) RemoteAddr() string             { return "127.0.0.1:1" }

func TestBindTypeCredentialsAndFullSequence(t *testing.T) {
	tests := []struct {
		name      string
		loginType sgip.LoginType
		password  string
		bound     bool
		want      sgip.RespStatus
	}{
		{name: "valid", loginType: sgip.SP_SMG, password: "sp-secret"},
		{name: "wrong direction", loginType: sgip.SMG_SP, password: "sp-secret", want: sgip.STAT_ERLGNTYPE},
		{name: "bad password", loginType: sgip.SP_SMG, password: "wrong", want: sgip.STAT_ILLLOGIN},
		{name: "duplicate", loginType: sgip.SP_SMG, password: "sp-secret", bound: true, want: sgip.STAT_RPTLOGIN},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &serverHandler{cfg: config{Credentials: exampleutil.Credentials{Account: "sp-demo", Password: "sp-secret"}}, logger: hlog.DefaultLogger()}
			request := sgip12.NewBind("sp-demo", tt.password, 3000000001, 5)
			request.Type = tt.loginType
			conn := &fakeConn{session: connectionSession{bound: tt.bound}}
			data, err := h.bind(context.Background(), conn, request)
			if (err == nil) != (tt.want == 0) {
				t.Fatalf("bind error=%v", err)
			}
			var response sgip12.BindResp
			if err := response.IDecode(data); err != nil {
				t.Fatal(err)
			}
			if response.Result != tt.want || response.Sequence != request.Sequence {
				t.Fatalf("bind response=%+v", response)
			}
		})
	}
}

func validSubmit() *sgip12.Submit {
	return &sgip12.Submit{
		Header:   sgip.Header{CommandID: sgip.SGIP_SUBMIT, Sequence: [3]uint32{3000000001, 1020123456, 7}},
		SpNumber: "1069", UserCount: 2, UserNumber: []string{"8613800138000", "8613900139000"}, ReportFlag: 1,
		MessageLength: 2, MessageContent: []byte("ok"),
	}
}

func TestSubmitReceiptBatchIsOwnedAndCorrelated(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	jobs := make(chan receiptJob, 1)
	h := &serverHandler{logger: hlog.DefaultLogger()}
	h.receipts = exampleutil.NewQueue(ctx, 1, 1, func(_ context.Context, j receiptJob) { jobs <- j })
	defer func() { cancel(); _ = h.receipts.Wait(context.Background()) }()
	request := validSubmit()
	data, err := h.submit(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	var response sgip12.SubmitResp
	if err := response.IDecode(data); err != nil {
		t.Fatal(err)
	}
	if response.Result != 0 || response.Sequence != request.Sequence {
		t.Fatal("submit response did not preserve full sequence")
	}
	request.UserNumber[0] = "changed after acceptance"
	select {
	case job := <-jobs:
		if len(job.recipients) != 2 || job.recipients[0] != "8613800138000" || job.submitSequence != response.Sequence {
			t.Fatal("queued batch aliased input or lost correlation")
		}
	case <-time.After(time.Second):
		t.Fatal("receipt batch not queued")
	}
}

func TestSubmitValidation(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*sgip12.Submit)
		want   sgip.RespStatus
	}{
		{name: "valid"},
		{name: "recipient count", mutate: func(p *sgip12.Submit) { p.UserCount = 101 }, want: sgip.STAT_ERARGFMT},
		{name: "missing country code", mutate: func(p *sgip12.Submit) { p.UserNumber[0] = "13800138000" }, want: sgip.STAT_ILLUSRNUM},
		{name: "unsupported billing mode", mutate: func(p *sgip12.Submit) { p.ReportFlag = 3 }, want: sgip.STAT_ERARGFMT},
		{name: "bad length", mutate: func(p *sgip12.Submit) { p.MessageLength = 3 }, want: sgip.STAT_ERLEN},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := validSubmit()
			if tt.mutate != nil {
				tt.mutate(p)
			}
			if got := validateSubmit(p); got != tt.want {
				t.Fatalf("status=%d want=%d", got, tt.want)
			}
		})
	}
}

type scriptWire struct {
	io.Reader
	io.Writer
}

func TestCallbackAcknowledgementUsesAllSequenceComponents(t *testing.T) {
	sequence := [3]uint32{2000000001, 1020123456, 9}
	tests := []struct {
		name     string
		request  protocol.PDU
		response protocol.PDU
		ok       bool
	}{
		{name: "bind", request: &sgip12.Bind{Header: sgip.Header{CommandID: sgip.SGIP_BIND, Sequence: sequence}, Type: sgip.SMG_SP}, response: &sgip12.BindResp{Header: sgip.Header{CommandID: sgip.SGIP_BIND_REP, Sequence: sequence}}, ok: true},
		{name: "report", request: &sgip12.Report{Header: sgip.Header{CommandID: sgip.SGIP_REPORT, Sequence: sequence}}, response: &sgip12.ReportResp{Header: sgip.Header{CommandID: sgip.SGIP_REPORT_REP, Sequence: sequence}}, ok: true},
		{name: "unbind", request: &sgip12.Unbind{Header: sgip.Header{CommandID: sgip.SGIP_UNBIND, Sequence: sequence}}, response: &sgip12.UnbindResp{Header: sgip.Header{CommandID: sgip.SGIP_UNBIND_REP, Sequence: sequence}}, ok: true},
		{name: "wrong node same counter", request: &sgip12.Report{Header: sgip.Header{CommandID: sgip.SGIP_REPORT, Sequence: sequence}}, response: &sgip12.ReportResp{Header: sgip.Header{CommandID: sgip.SGIP_REPORT_REP, Sequence: [3]uint32{123, sequence[1], sequence[2]}}}},
		{name: "rejected report", request: &sgip12.Report{Header: sgip.Header{CommandID: sgip.SGIP_REPORT, Sequence: sequence}}, response: &sgip12.ReportResp{Header: sgip.Header{CommandID: sgip.SGIP_REPORT_REP, Sequence: sequence}, Result: sgip.STAT_ERSEQ}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input, output bytes.Buffer
			if err := exampleutil.Write(&input, tt.response); err != nil {
				t.Fatal(err)
			}
			if err := expectResponse(scriptWire{Reader: &input, Writer: &output}, tt.request); (err == nil) != tt.ok {
				t.Fatalf("response validation=%v", err)
			}
		})
	}
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

func TestConfiguration(t *testing.T) {
	if _, err := parseConfig(nil, testEnv, io.Discard); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		args []string
	}{
		{name: "node overflow", args: []string{"-node-id=4294967296"}},
		{name: "callback port", args: []string{"-sp-address=127.0.0.1:0"}},
		{name: "callback timeout", args: []string{"-callback-timeout=0"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := parseConfig(tt.args, testEnv, io.Discard); err == nil {
				t.Fatal("invalid configuration accepted")
			}
		})
	}
	if realMain([]string{"-help"}, testEnv, io.Discard) != 0 {
		t.Fatal("help failed")
	}
	if realMain([]string{"-receipt-workers=0"}, testEnv, io.Discard) != 2 {
		t.Fatal("configuration exit code")
	}
}
