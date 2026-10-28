package main

import (
	"context"
	"encoding/binary"
	"testing"
	"time"

	"github.com/cloudwego/netpoll"

	smslog "github.com/hujm2023/go-sms-protocol/logger"
	"github.com/hujm2023/go-sms-protocol/nioserver"
	"github.com/hujm2023/go-sms-protocol/smgp"
	"github.com/hujm2023/go-sms-protocol/smgp/smgp30"
)

func validSubmit() *smgp30.Submit {
	return &smgp30.Submit{
		Header:  smgp.NewHeader(0, smgp.CommandSubmit, 9),
		MsgType: smgp.MT, NeedReport: smgp.NEED_REPORT, SrcTermID: "118900001",
		DestTermIDCount: 1, DestTermID: []string{"13800138000"},
		MsgLength: 2, MsgContent: []byte("ok"),
	}
}

func TestUnpackRejectsInvalidFrameSizes(t *testing.T) {
	valid, err := (&smgp30.ActiveTest{Header: smgp.NewHeader(0, smgp.CommandActiveTest, 7)}).IEncode()
	if err != nil {
		t.Fatal(err)
	}
	extra := append(append([]byte(nil), valid...), 0)
	binary.BigEndian.PutUint32(extra, uint32(len(extra)))
	oversized := make([]byte, 4)
	binary.BigEndian.PutUint32(oversized, maxSMGP30FrameSize+1)
	tests := []struct {
		name string
		data []byte
		ok   bool
	}{
		{name: "valid", data: valid, ok: true},
		{name: "zero size", data: make([]byte, 4)},
		{name: "oversized", data: oversized},
		{name: "extra body byte", data: extra},
		{name: "truncated header", data: valid[:8]},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := netpoll.NewLinkBuffer()
			defer reader.Close()
			if _, err := reader.WriteBinary(tt.data); err != nil {
				t.Fatal(err)
			}
			if err := reader.Flush(); err != nil {
				t.Fatal(err)
			}
			pdu, err := newUnpackFunc(maxSMGP30FrameSize)(context.Background(), reader)
			if (err == nil) != tt.ok {
				t.Fatalf("unpack error = %v, want success=%t", err, tt.ok)
			}
			if tt.ok && pdu.GetSequenceID() != 7 {
				t.Fatal("unpack lost the request sequence")
			}
		})
	}
}

func TestSubmitValidationAndLoginMode(t *testing.T) {
	tests := []struct {
		name   string
		mode   uint8
		mutate func(*smgp30.Submit)
		want   uint32
	}{
		{name: "transmit", mode: 2},
		{name: "receive only", mode: 1, want: 10},
		{name: "send only cannot receive reports", mode: 0, want: 10},
		{name: "send without report", mode: 0, mutate: func(p *smgp30.Submit) { p.NeedReport = 0 }},
		{name: "invalid type", mode: 2, mutate: func(p *smgp30.Submit) { p.MsgType = 0 }, want: 30},
		{name: "invalid priority", mode: 2, mutate: func(p *smgp30.Submit) { p.Priority = 10 }, want: 31},
		{name: "invalid report flag", mode: 2, mutate: func(p *smgp30.Submit) { p.NeedReport = 2 }, want: 10},
		{name: "no recipients", mode: 2, mutate: func(p *smgp30.Submit) { p.DestTermIDCount = 0 }, want: 47},
		{name: "empty recipient", mode: 2, mutate: func(p *smgp30.Submit) { p.DestTermID[0] = "" }, want: 47},
		{name: "missing source", mode: 2, mutate: func(p *smgp30.Submit) { p.SrcTermID = "" }, want: 46},
		{name: "bad encoding", mode: 2, mutate: func(p *smgp30.Submit) { p.MsgFormat = 3 }, want: 34},
		{name: "too long", mode: 2, mutate: func(p *smgp30.Submit) { p.MsgContent = make([]byte, 141); p.MsgLength = 141 }, want: 36},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn := newFakeSMSConn()
			conn.SetBizData(connectionSession{account: "900001", loginMode: tt.mode, isAuthenticated: true})
			d := &receiptDispatcher{ctx: context.Background(), jobs: make(chan receiptJob, 100)}
			h := &serverHandler{receipts: d, logger: smslog.DefaultLogger(), now: time.Now}
			request := validSubmit()
			if tt.mutate != nil {
				tt.mutate(request)
			}
			data, err := h.handleSubmit(context.Background(), conn, request)
			if err != nil {
				t.Fatal(err)
			}
			var response smgp30.SubmitResp
			if err := response.IDecode(data); err != nil {
				t.Fatal(err)
			}
			if response.Status != tt.want || response.GetSequenceID() != 9 {
				t.Fatalf("submit result = %d sequence = %d, want %d/9", response.Status, response.GetSequenceID(), tt.want)
			}
			if tt.want != 0 && len(d.jobs) != 0 {
				t.Fatal("rejected submit scheduled receipts")
			}
		})
	}
}

func TestSubmitBackpressureAndReportCorrelation(t *testing.T) {
	conn := newFakeSMSConn()
	conn.SetBizData(connectionSession{account: "900001", loginMode: smgp.TRANSMIT_MODE, isAuthenticated: true})
	d := &receiptDispatcher{ctx: context.Background(), jobs: make(chan receiptJob, 1)}
	h := &serverHandler{receipts: d, logger: smslog.DefaultLogger(), now: time.Now}
	data, err := h.handleSubmit(context.Background(), conn, validSubmit())
	if err != nil {
		t.Fatal(err)
	}
	var accepted smgp30.SubmitResp
	if err := accepted.IDecode(data); err != nil {
		t.Fatal(err)
	}
	data, err = h.handleSubmit(context.Background(), conn, validSubmit())
	if err != nil {
		t.Fatal(err)
	}
	var rejected smgp30.SubmitResp
	if err := rejected.IDecode(data); err != nil {
		t.Fatal(err)
	}
	if rejected.Status != 1 {
		t.Fatalf("full receipt queue returned %d, want system busy (1)", rejected.Status)
	}
	job := <-d.jobs
	if accepted.Status != 0 || job.messageID != accepted.MsgID || job.conn != conn {
		t.Fatal("accepted receipt lost the submit ID or submitting connection")
	}
}

func TestLoginDoesNotAuthenticateRejectedConnection(t *testing.T) {
	conn := newFakeSMSConn()
	manager := &heartbeatManager{
		ctx: context.Background(), connections: make(map[nioserver.ISMSConn[connectionSession]]struct{}),
	}
	h := &serverHandler{
		account: "900001", password: "test-secret", authSkew: time.Minute,
		heartbeats: manager, logger: smslog.DefaultLogger(), now: time.Now,
	}
	login := smgp30.NewLogin(h.account, "wrong-password", 1)
	data, err := h.handleLogin(context.Background(), conn, login)
	if err != errLoginRejected || conn.GetBizData().isAuthenticated {
		t.Fatal("bad login changed authentication state")
	}
	var response smgp30.LoginResp
	if err := response.IDecode(data); err != nil {
		t.Fatal(err)
	}
	if response.Status != 21 {
		t.Fatalf("bad login status = %d, want authentication error (21)", response.Status)
	}
	login = smgp30.NewLogin(h.account, h.password, 2)
	if _, err := h.handleLogin(context.Background(), conn, login); err != nil {
		t.Fatal(err)
	}
	if !conn.GetBizData().isAuthenticated || len(manager.connections) != 1 {
		t.Fatal("valid login did not authenticate and register heartbeat")
	}
	if _, err := h.handleLogin(context.Background(), conn, login); err != errLoginRejected {
		t.Fatal("duplicate login accepted")
	}
}
