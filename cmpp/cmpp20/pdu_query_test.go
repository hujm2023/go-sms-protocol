package cmpp20

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/hujm2023/go-sms-protocol/cmpp"
)

func decodeQueryFixture(t *testing.T, value string) []byte {
	t.Helper()
	data, err := hex.DecodeString(value)
	if err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	return data
}

func TestPduQuery_RoundTripAndDispatcher(t *testing.T) {
	raw := decodeQueryFixture(t, "0000002700000006000001233230323630383235017376632d616c706861007273760000000000")

	var query PduQuery
	if err := query.IDecode(raw); err != nil {
		t.Fatalf("IDecode() error = %v", err)
	}
	if query.CommandID != cmpp.CommandQuery || query.SequenceID != 0x123 {
		t.Fatalf("header = %#v, want command %v and sequence 0x123", query.Header, cmpp.CommandQuery)
	}
	if query.Time != "20260825" || query.QueryType != 1 || query.QueryCode != "svc-alpha" || query.Reserve != "rsv" {
		t.Fatalf("decoded query fields = %#v", query)
	}

	encoded, err := query.IEncode()
	if err != nil {
		t.Fatalf("IEncode() error = %v", err)
	}
	if !bytes.Equal(encoded, raw) {
		t.Fatalf("IEncode() = %x, want %x", encoded, raw)
	}

	pdu, err := DecodeCMPP20(raw)
	if err != nil {
		t.Fatalf("DecodeCMPP20() error = %v", err)
	}
	decoded, ok := pdu.(*PduQuery)
	if !ok {
		t.Fatalf("DecodeCMPP20() type = %T, want *PduQuery", pdu)
	}
	if decoded.GetCommand() != cmpp.CommandQuery || decoded.GetSequenceID() != query.GetSequenceID() {
		t.Fatalf("dispatcher result = %#v", decoded)
	}
	response := query.GenEmptyResponse()
	if response.GetCommand() != cmpp.CommandQueryResp || response.GetSequenceID() != query.GetSequenceID() {
		t.Fatalf("GenEmptyResponse() = %#v", response)
	}
}

func TestPduQueryResp_RoundTripAndDispatcher(t *testing.T) {
	raw := decodeQueryFixture(t, "0000003f80000006000004563230323630383235017376632d616c706861000102030411121314212223243132333441424344515253546162636471727374")

	var response PduQueryResp
	if err := response.IDecode(raw); err != nil {
		t.Fatalf("IDecode() error = %v", err)
	}
	want := PduQueryResp{
		Time:      "20260825",
		QueryType: 1,
		QueryCode: "svc-alpha",
		MtTLMsg:   0x01020304,
		MtTlUsr:   0x11121314,
		MtScs:     0x21222324,
		MtWT:      0x31323334,
		MtFL:      0x41424344,
		MoScs:     0x51525354,
		MoWT:      0x61626364,
		MoFL:      0x71727374,
	}
	if response.Time != want.Time || response.QueryType != want.QueryType || response.QueryCode != want.QueryCode ||
		response.MtTLMsg != want.MtTLMsg || response.MtTlUsr != want.MtTlUsr || response.MtScs != want.MtScs ||
		response.MtWT != want.MtWT || response.MtFL != want.MtFL || response.MoScs != want.MoScs ||
		response.MoWT != want.MoWT || response.MoFL != want.MoFL {
		t.Fatalf("decoded query response = %#v, want %#v", response, want)
	}

	encoded, err := response.IEncode()
	if err != nil {
		t.Fatalf("IEncode() error = %v", err)
	}
	if !bytes.Equal(encoded, raw) {
		t.Fatalf("IEncode() = %x, want %x", encoded, raw)
	}

	pdu, err := DecodeCMPP20(raw)
	if err != nil {
		t.Fatalf("DecodeCMPP20() error = %v", err)
	}
	decoded, ok := pdu.(*PduQueryResp)
	if !ok {
		t.Fatalf("DecodeCMPP20() type = %T, want *PduQueryResp", pdu)
	}
	if decoded.MoScs != want.MoScs || decoded.MoWT != want.MoWT || decoded.MoFL != want.MoFL {
		t.Fatalf("dispatcher lost MO counters = %#v", decoded)
	}
}
