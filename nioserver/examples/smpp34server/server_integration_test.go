//go:build integration

package main

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	protocol "github.com/hujm2023/go-sms-protocol"
	smslog "github.com/hujm2023/go-sms-protocol/logger"
	"github.com/hujm2023/go-sms-protocol/nioserver/examples/internal/exampleutil"
	"github.com/hujm2023/go-sms-protocol/smpp"
	"github.com/hujm2023/go-sms-protocol/smpp/smpp34"
)

func startNetworkServer(t *testing.T) string {
	t.Helper()
	env := func(name string) (string, bool) {
		if name == "SMPP34_ACCOUNT" {
			return "demo", true
		}
		return "secret34", true
	}
	cfg, err := parseConfig([]string{"-listen=127.0.0.1:0", "-heartbeat-interval=20ms", "-receipt-min-delay=1ms", "-receipt-max-delay=1ms"}, env, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	result := make(chan error, 1)
	ready := make(chan net.Addr, 1)
	go func() { result <- run(ctx, cfg, smslog.DefaultLogger(), func(address net.Addr) { ready <- address }) }()
	t.Cleanup(func() {
		cancel()
		select {
		case err := <-result:
			if err != nil {
				t.Error(err)
			}
		case <-time.After(2 * time.Second):
			t.Error("server did not stop")
		}
	})
	select {
	case address := <-ready:
		return address.String()
	case <-time.After(time.Second):
		t.Fatal("server did not bind")
	}
	return ""
}

func dialNetwork(t *testing.T, address string) net.Conn {
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

func sendNetwork(t *testing.T, conn net.Conn, pdu protocol.PDU) {
	t.Helper()
	if err := exampleutil.Write(conn, pdu); err != nil {
		t.Fatal(err)
	}
}
func receiveNetwork(t *testing.T, conn net.Conn) protocol.PDU {
	t.Helper()
	pdu, err := exampleutil.Read(conn, 16, 4096, smpp34.DecodeSMPP34)
	if err != nil {
		t.Fatal(err)
	}
	return pdu
}

func TestTCPReceiptHeartbeatAndUnbind(t *testing.T) {
	address := startNetworkServer(t)
	conn := dialNetwork(t, address)
	sendNetwork(t, conn, &smpp34.Bind{Header: smpp.Header{ID: smpp.BIND_TRANSCEIVER, Sequence: 1}, SystemID: "demo", Password: "secret34", InterfaceVersion: 0x34})
	if p, ok := receiveNetwork(t, conn).(*smpp34.BindResp); !ok || p.Status != 0 {
		t.Fatal("bind failed")
	}
	sendNetwork(t, conn, &smpp34.SubmitSm{Header: smpp.Header{ID: smpp.SUBMIT_SM, Sequence: 2}, SourceAddr: "1069", DestinationAddr: "13800138000", RegisteredDelivery: 1})
	sendNetwork(t, conn, &smpp34.EnquireLink{Header: smpp.Header{ID: smpp.ENQUIRE_LINK, Sequence: 3}})
	var submitID, reportID string
	var clientHeartbeat, serverHeartbeat bool
	for submitID == "" || reportID == "" || !clientHeartbeat || !serverHeartbeat {
		switch p := receiveNetwork(t, conn).(type) {
		case *smpp34.SubmitSmResp:
			if p.Status != 0 {
				t.Fatalf("submit status=%d", p.Status)
			}
			submitID = p.MessageID
		case *smpp34.DeliverSm:
			var report smpp34.DeliveryReceipt
			if err := report.IDecode(p.ShortMessage); err != nil {
				t.Fatal(err)
			}
			if report.Stat != "DELIVRD" || p.ESMClass != 4 {
				t.Fatal("wrong receipt marker/state")
			}
			reportID = report.ID
			sendNetwork(t, conn, p.GenEmptyResponse())
		case *smpp34.EnquireLink:
			serverHeartbeat = true
			sendNetwork(t, conn, p.GenEmptyResponse())
		case *smpp34.EnquireLinkResp:
			clientHeartbeat = p.Sequence == 3 && p.Status == 0
		default:
			t.Fatalf("unexpected response %T", p)
		}
	}
	if submitID != reportID {
		t.Fatal("receipt did not correlate with submit_sm_resp")
	}
	sendNetwork(t, conn, &smpp34.Unbind{Header: smpp.Header{ID: smpp.UNBIND, Sequence: 4}})
	for {
		pdu := receiveNetwork(t, conn)
		if p, ok := pdu.(*smpp34.EnquireLink); ok {
			sendNetwork(t, conn, p.GenEmptyResponse())
			continue
		}
		if p, ok := pdu.(*smpp34.UnBindResp); !ok || p.Sequence != 4 {
			t.Fatal("missing unbind response")
		}
		break
	}
	var b [1]byte
	if _, err := conn.Read(b[:]); err != io.EOF {
		t.Fatalf("expected close after unbind: %v", err)
	}
}

func TestTCPRejectedBindFlushesResponse(t *testing.T) {
	address := startNetworkServer(t)
	conn := dialNetwork(t, address)
	sendNetwork(t, conn, &smpp34.Bind{Header: smpp.Header{ID: smpp.BIND_TRANSCEIVER, Sequence: 1}, SystemID: "demo", Password: "wrong", InterfaceVersion: 0x34})
	if p, ok := receiveNetwork(t, conn).(*smpp34.BindResp); !ok || p.Status != smpp.ESME_RINVPASWD {
		t.Fatal("wrong password not rejected")
	}
	var b [1]byte
	if _, err := conn.Read(b[:]); err != io.EOF {
		t.Fatalf("rejected bind connection not closed: %v", err)
	}
}
