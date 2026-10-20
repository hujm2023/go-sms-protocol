package main

import (
	"bytes"
	"io"
	"strings"
	"testing"

	protocol "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/nioserver/examples/internal/exampleutil"
	"github.com/hujm2023/go-sms-protocol/smpp"
	"github.com/hujm2023/go-sms-protocol/smpp/smpp34"
)

type sessionWire struct {
	io.Reader
	io.Writer
}

func TestSessionAcknowledgesOutOfOrderReceiptAndBothHeartbeats(t *testing.T) {
	content, err := (&smpp34.DeliveryReceipt{ID: "42", Stat: "DELIVRD", Err: "000"}).IEncode()
	if err != nil {
		t.Fatal(err)
	}
	responses := []protocol.PDU{
		&smpp34.BindResp{Header: smpp.Header{ID: smpp.BIND_TRANSCEIVER_RESP, Sequence: 1}, SystemID: "smsc"},
		&smpp34.DeliverSm{Header: smpp.Header{ID: smpp.DELIVER_SM, Sequence: 100}, ESMClass: 4, ShortMessage: content, SmLength: uint8(len(content))},
		&smpp34.EnquireLink{Header: smpp.Header{ID: smpp.ENQUIRE_LINK, Sequence: 101}},
		&smpp34.SubmitSmResp{Header: smpp.Header{ID: smpp.SUBMIT_SM_RESP, Sequence: 2}, MessageID: "42"},
		&smpp34.EnquireLinkResp{Header: smpp.Header{ID: smpp.ENQUIRE_LINK_RESP, Sequence: 3}},
		&smpp34.UnBindResp{Header: smpp.Header{ID: smpp.UNBIND_RESP, Sequence: 4}},
	}
	var input, output, log bytes.Buffer
	for _, pdu := range responses {
		if err := exampleutil.Write(&input, pdu); err != nil {
			t.Fatal(err)
		}
	}
	cfg := config{Credentials: exampleutil.Credentials{Account: "demo", Password: "secret34"}, source: "1069", destination: "13800138000", message: "hello"}
	if err := runSession(sessionWire{Reader: &input, Writer: &output}, cfg, &log); err != nil {
		t.Fatal(err)
	}
	var receiptAck, heartbeatAck bool
	for output.Len() > 0 {
		pdu, err := exampleutil.Read(&output, 16, 4096, smpp34.DecodeSMPP34)
		if err != nil {
			t.Fatal(err)
		}
		switch p := pdu.(type) {
		case *smpp34.DeliverSmResp:
			receiptAck = p.Sequence == 100 && p.Status == 0 && p.MessageID == ""
		case *smpp34.EnquireLinkResp:
			heartbeatAck = p.Sequence == 101 && p.Status == 0
		}
	}
	if !receiptAck || !heartbeatAck || !strings.Contains(log.String(), "unbind acknowledged") {
		t.Fatal("session did not acknowledge all required PDUs")
	}
}

func TestSessionRejectsFailedBind(t *testing.T) {
	var input, output bytes.Buffer
	if err := exampleutil.Write(&input, &smpp34.BindResp{Header: smpp.Header{ID: smpp.BIND_TRANSCEIVER_RESP, Sequence: 1, Status: smpp.ESME_RINVPASWD}}); err != nil {
		t.Fatal(err)
	}
	if err := runSession(sessionWire{Reader: &input, Writer: &output}, config{}, io.Discard); err == nil {
		t.Fatal("failed bind accepted")
	}
}

func TestClientConfiguration(t *testing.T) {
	env := func(name string) (string, bool) {
		if name == "SMPP34_ACCOUNT" {
			return "demo", true
		}
		return "secret34", true
	}
	if _, err := parseConfig(nil, env, io.Discard); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		args []string
	}{
		{name: "deadline", args: []string{"-timeout=0"}},
		{name: "non-ASCII", args: []string{"-message=中文"}},
		{name: "long address", args: []string{"-source=" + strings.Repeat("1", 21)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := parseConfig(tt.args, env, io.Discard); err == nil {
				t.Fatal("invalid arguments accepted")
			}
		})
	}
	if realMain([]string{"-help"}, env, io.Discard, io.Discard) != 0 {
		t.Fatal("help failed")
	}
	if realMain([]string{"-timeout=0"}, env, io.Discard, io.Discard) != 2 {
		t.Fatal("configuration exit code")
	}
}
