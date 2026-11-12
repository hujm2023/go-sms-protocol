package main

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	smslog "github.com/hujm2023/go-sms-protocol/logger"
	"github.com/hujm2023/go-sms-protocol/nioserver"
	"github.com/hujm2023/go-sms-protocol/nioserver/examples/internal/exampleutil"
	"github.com/hujm2023/go-sms-protocol/smpp"
	"github.com/hujm2023/go-sms-protocol/smpp/smpp34"
)

type fakeConn struct {
	nioserver.ISMSConn[connectionSession]
	session  connectionSession
	sequence uint32
	missed   int
	closed   bool
	writes   [][]byte
}

func (c *fakeConn) GetBizData() connectionSession  { return c.session }
func (c *fakeConn) SetBizData(s connectionSession) { c.session = s }
func (c *fakeConn) RemoteAddr() string             { return "127.0.0.1:1" }
func (c *fakeConn) NextSequenceID() uint32         { c.sequence++; return c.sequence }
func (c *fakeConn) NoActiveTestCount() int         { return c.missed }
func (c *fakeConn) OnSendActiveTest()              { c.missed++ }
func (c *fakeConn) OnReceiveActiveTest()           { c.missed = 0 }
func (c *fakeConn) Close() error                   { c.closed = true; return nil }
func (c *fakeConn) Write(_ context.Context, data []byte) error {
	c.writes = append(c.writes, append([]byte(nil), data...))
	return nil
}

func newHandler() *serverHandler {
	return &serverHandler{
		cfg:        config{Credentials: exampleutil.Credentials{Account: "demo", Password: "secret34"}},
		logger:     smslog.DefaultLogger(),
		heartbeats: &heartbeatManager{ctx: context.Background(), connections: make(map[nioserver.ISMSConn[connectionSession]]struct{})},
	}
}

func TestBindCredentialsAndModes(t *testing.T) {
	tests := []struct {
		name              string
		account, password string
		mode              smpp.CMDId
		bound             bool
		want              smpp.CMDStatus
	}{
		{name: "transceiver", account: "demo", password: "secret34", mode: smpp.BIND_TRANSCEIVER},
		{name: "receiver", account: "demo", password: "secret34", mode: smpp.BIND_RECEIVER},
		{name: "transmitter", account: "demo", password: "secret34", mode: smpp.BIND_TRANSMITTER},
		{name: "wrong account", account: "other", password: "secret34", mode: smpp.BIND_TRANSCEIVER, want: smpp.ESME_RINVSYSID},
		{name: "wrong password", account: "demo", password: "wrong", mode: smpp.BIND_TRANSCEIVER, want: smpp.ESME_RINVPASWD},
		{name: "duplicate", account: "demo", password: "secret34", mode: smpp.BIND_TRANSCEIVER, bound: true, want: smpp.ESME_RALYBND},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newHandler()
			conn := &fakeConn{session: connectionSession{bound: tt.bound}}
			request := &smpp34.Bind{Header: smpp.Header{ID: tt.mode, Sequence: 42}, SystemID: tt.account, Password: tt.password, InterfaceVersion: 0x34}
			data, err := h.bind(context.Background(), conn, request)
			if (err == nil) != (tt.want == 0) {
				t.Fatalf("bind error=%v", err)
			}
			var response smpp34.BindResp
			if err := response.IDecode(data); err != nil {
				t.Fatal(err)
			}
			if response.Status != tt.want || response.Sequence != 42 {
				t.Fatalf("bind response=%+v", response)
			}
			if tt.want == 0 && !conn.session.bound {
				t.Fatal("valid bind did not establish session")
			}
			if tt.want != 0 && !tt.bound && conn.session.bound {
				t.Fatal("rejected bind authenticated session")
			}
		})
	}
}

func TestSubmitHonorsBindModeAndReceiptFlags(t *testing.T) {
	tests := []struct {
		name       string
		mode       smpp.CMDId
		registered uint8
		want       smpp.CMDStatus
	}{
		{name: "transceiver report", mode: smpp.BIND_TRANSCEIVER, registered: 1},
		{name: "failure-only receipt", mode: smpp.BIND_TRANSCEIVER, registered: 2},
		{name: "transmitter no report", mode: smpp.BIND_TRANSMITTER},
		{name: "receiver cannot submit", mode: smpp.BIND_RECEIVER, want: smpp.ESME_RINVBNDSTS},
		{name: "transmitter cannot receive report", mode: smpp.BIND_TRANSMITTER, registered: 1, want: smpp.ESME_RINVBNDSTS},
		{name: "unsupported acknowledgement flags", mode: smpp.BIND_TRANSCEIVER, registered: 4, want: smpp.ESME_RINVREGDLVFLG},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			jobs := make(chan receiptJob, 1)
			h := newHandler()
			h.receipts = exampleutil.NewQueue(ctx, 1, 1, func(_ context.Context, j receiptJob) { jobs <- j })
			defer func() {
				cancel()
				if err := h.receipts.Wait(context.Background()); err != nil {
					t.Error(err)
				}
			}()
			conn := &fakeConn{session: connectionSession{bound: true, mode: tt.mode}}
			p := &smpp34.SubmitSm{Header: smpp.Header{ID: smpp.SUBMIT_SM, Sequence: 7}, SourceAddr: "1069", DestinationAddr: "13800138000", RegisteredDelivery: tt.registered}
			data, err := h.submit(context.Background(), conn, p)
			if err != nil {
				t.Fatal(err)
			}
			var response smpp34.SubmitSmResp
			if err := response.IDecode(data); err != nil {
				t.Fatal(err)
			}
			if response.Status != tt.want || response.Sequence != 7 {
				t.Fatalf("submit response=%+v", response)
			}
			if tt.want == 0 && tt.registered == 1 {
				select {
				case job := <-jobs:
					if job.messageID != response.MessageID || job.conn != conn {
						t.Fatal("receipt lost original message ID or connection")
					}
				case <-time.After(time.Second):
					t.Fatal("no receipt queued")
				}
			}
		})
	}
}

func TestSubmitChecksReceiptSizeBeforeAccepting(t *testing.T) {
	template, err := buildReceipt(receiptJob{messageID: "1", destination: "13800138000"}, time.Now(), 1)
	if err != nil {
		t.Fatal(err)
	}
	wire, err := template.IEncode()
	if err != nil {
		t.Fatal(err)
	}
	maxSource := smpp.MAX_PDU_SIZE - len(wire)
	for _, test := range []struct {
		name         string
		sourceLength int
		want         smpp.CMDStatus
	}{
		{name: "receipt fits exactly", sourceLength: maxSource},
		{name: "receipt one byte too large", sourceLength: maxSource + 1, want: smpp.ESME_RINVMSGLEN},
		{name: "4000-byte source", sourceLength: 4000, want: smpp.ESME_RINVMSGLEN},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			jobs := make(chan receiptJob, 1)
			handler := newHandler()
			handler.receipts = exampleutil.NewQueue(ctx, 1, 1, func(_ context.Context, job receiptJob) { jobs <- job })
			defer func() {
				cancel()
				if err := handler.receipts.Wait(context.Background()); err != nil {
					t.Error(err)
				}
			}()
			conn := &fakeConn{session: connectionSession{bound: true, mode: smpp.BIND_TRANSCEIVER}}
			request := &smpp34.SubmitSm{
				Header:          smpp.Header{ID: smpp.SUBMIT_SM, Sequence: 7},
				SourceAddr:      strings.Repeat("1", test.sourceLength),
				DestinationAddr: "13800138000", RegisteredDelivery: 1,
			}
			if _, err := request.IEncode(); err != nil {
				t.Fatalf("submit must fit: %v", err)
			}
			wire, err := handler.submit(ctx, conn, request)
			if err != nil {
				t.Fatal(err)
			}
			var response smpp34.SubmitSmResp
			if err := response.IDecode(wire); err != nil {
				t.Fatal(err)
			}
			if response.Status != test.want {
				t.Fatalf("status=%d, want %d", response.Status, test.want)
			}
			if test.want != smpp.ESME_ROK {
				if response.MessageID != "" {
					t.Fatal("rejected submit has a message ID")
				}
				cancel()
				if err := handler.receipts.Wait(context.Background()); err != nil {
					t.Fatal(err)
				}
				if len(jobs) != 0 {
					t.Fatal("oversized receipt was queued")
				}
				return
			}
			select {
			case job := <-jobs:
				receipt, err := buildReceipt(job, time.Now(), 42)
				if err != nil {
					t.Fatal(err)
				}
				wire, err := receipt.IEncode()
				if err != nil {
					t.Fatal(err)
				}
				if len(wire) != smpp.MAX_PDU_SIZE {
					t.Fatalf("receipt length=%d", len(wire))
				}
			case <-time.After(time.Second):
				t.Fatal("accepted submit has no receipt")
			}
		})
	}
}

func TestReceiptWireFields(t *testing.T) {
	job := receiptJob{messageID: "12345", source: "1069", destination: "13800138000", sourceTON: 5, destinationTON: 1, submitted: time.Now()}
	p, err := buildReceipt(job, time.Now(), 10)
	if err != nil {
		t.Fatal(err)
	}
	data, err := p.IEncode()
	if err != nil {
		t.Fatal(err)
	}
	var decoded smpp34.DeliverSm
	if err := decoded.IDecode(data); err != nil {
		t.Fatal(err)
	}
	var receipt smpp34.DeliveryReceipt
	if err := receipt.IDecode(decoded.ShortMessage); err != nil {
		t.Fatal(err)
	}
	if decoded.ESMClass != 4 || receipt.ID != job.messageID || receipt.Stat != "DELIVRD" || decoded.SourceAddr != job.destination || decoded.DestinationAddr != job.source {
		t.Fatal("incorrect receipt identifiers, direction or marker")
	}
	ids := decoded.OptionalTLVs.All(smpp.RECEIPTED_MESSAGE_ID)
	states := decoded.OptionalTLVs.All(smpp.DR_MESSAGE_STATE)
	if len(ids) != 1 || string(ids[0].ValueBytes) != "12345\x00" || len(states) != 1 || string(states[0].ValueBytes) != "\x02" {
		t.Fatal("receipt TLVs did not match the text report")
	}
}

func TestEnquireLinkSequenceAndTimeout(t *testing.T) {
	manager := &heartbeatManager{ctx: context.Background(), logger: smslog.DefaultLogger(), maxMissed: 3, writeTimeout: time.Second, connections: make(map[nioserver.ISMSConn[connectionSession]]struct{})}
	conn := &fakeConn{sequence: smpp.SEQUENCE_NUM_END}
	manager.checkConnection(conn)
	var heartbeat smpp34.EnquireLink
	if err := heartbeat.IDecode(conn.writes[0]); err != nil {
		t.Fatal(err)
	}
	if heartbeat.Sequence != 1 || conn.missed != 1 {
		t.Fatal("heartbeat did not wrap within SMPP's positive 31-bit sequence range")
	}
	conn.missed = 3
	manager.checkConnection(conn)
	if !conn.closed {
		t.Fatal("missed heartbeats did not close connection")
	}
}

func TestConfiguration(t *testing.T) {
	env := func(name string) (string, bool) {
		if name == "SMPP34_ACCOUNT" {
			return "demo", true
		}
		return "secret34", true
	}
	if _, err := parseConfig(nil, env, io.Discard); err != nil {
		t.Fatal(err)
	}
	if _, err := parseConfig([]string{"-heartbeat-interval=0"}, env, io.Discard); err == nil {
		t.Fatal("invalid heartbeat accepted")
	}
	if realMain([]string{"-help"}, env, io.Discard) != 0 {
		t.Fatal("help failed")
	}
	if realMain([]string{"-receipt-queue=0"}, env, io.Discard) != 2 {
		t.Fatal("configuration exit code")
	}
}
