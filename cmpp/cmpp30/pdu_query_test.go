package cmpp30

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/hujm2023/go-sms-protocol/cmpp"
)

func decodeCMPP30Fixture(t *testing.T, value string) []byte {
	t.Helper()
	data, err := hex.DecodeString(value)
	if err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	return data
}

func TestQuery_RoundTripAndDispatcher(t *testing.T) {
	raw := decodeCMPP30Fixture(t, "0000002700000006000002013230323630383235017376632d616c706861007273760000000000")

	var query Query
	if err := query.IDecode(raw); err != nil {
		t.Fatalf("IDecode() error = %v", err)
	}
	if query.CommandID != cmpp.CommandQuery || query.SequenceID != 0x201 || query.Time != "20260825" ||
		query.QueryType != 1 || query.QueryCode != "svc-alpha" || query.Reserve != "rsv" {
		t.Fatalf("decoded query = %#v", query)
	}
	encoded, err := query.IEncode()
	if err != nil {
		t.Fatalf("IEncode() error = %v", err)
	}
	if !bytes.Equal(encoded, raw) {
		t.Fatalf("IEncode() = %x, want %x", encoded, raw)
	}

	pdu, err := DecodeCMPP30(raw)
	if err != nil {
		t.Fatalf("DecodeCMPP30() error = %v", err)
	}
	decoded, ok := pdu.(*Query)
	if !ok || decoded.GetSequenceID() != query.GetSequenceID() {
		t.Fatalf("dispatcher result = %T %#v", pdu, pdu)
	}
	response := query.GenEmptyResponse()
	if response.GetCommand() != cmpp.CommandQueryResp || response.GetSequenceID() != query.GetSequenceID() {
		t.Fatalf("GenEmptyResponse() = %#v", response)
	}
}

func TestQueryResp_RoundTripAndDispatcher(t *testing.T) {
	raw := decodeCMPP30Fixture(t, "0000003f80000006000002023230323630383235017376632d616c706861000102030411121314212223243132333441424344515253546162636471727374")

	var response QueryResp
	if err := response.IDecode(raw); err != nil {
		t.Fatalf("IDecode() error = %v", err)
	}
	want := [8]uint32{0x01020304, 0x11121314, 0x21222324, 0x31323334, 0x41424344, 0x51525354, 0x61626364, 0x71727374}
	got := [8]uint32{response.MtTLMsg, response.MtTlUsr, response.MtScs, response.MtWT, response.MtFL, response.MoScs, response.MoWT, response.MoFL}
	if response.Time != "20260825" || response.QueryType != 1 || response.QueryCode != "svc-alpha" || got != want {
		t.Fatalf("decoded query response = %#v, counters = %#v", response, got)
	}
	encoded, err := response.IEncode()
	if err != nil {
		t.Fatalf("IEncode() error = %v", err)
	}
	if !bytes.Equal(encoded, raw) {
		t.Fatalf("IEncode() = %x, want %x", encoded, raw)
	}

	pdu, err := DecodeCMPP30(raw)
	if err != nil {
		t.Fatalf("DecodeCMPP30() error = %v", err)
	}
	decoded, ok := pdu.(*QueryResp)
	if !ok || decoded.MoFL != want[7] {
		t.Fatalf("dispatcher result = %T %#v", pdu, pdu)
	}
}
