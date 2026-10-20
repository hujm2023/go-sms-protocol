package main

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/hujm2023/go-sms-protocol/nioserver"
	"github.com/hujm2023/go-sms-protocol/nioserver/examples/internal/exampleutil"
	"github.com/hujm2023/go-sms-protocol/smpp"
	"github.com/hujm2023/go-sms-protocol/smpp/smpp34"
	"github.com/hujm2023/hlog"
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
		logger:     hlog.DefaultLogger(),
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
	manager := &heartbeatManager{ctx: context.Background(), logger: hlog.DefaultLogger(), maxMissed: 3, writeTimeout: time.Second, connections: make(map[nioserver.ISMSConn[connectionSession]]struct{})}
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
