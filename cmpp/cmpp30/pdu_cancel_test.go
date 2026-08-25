package cmpp30

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/hujm2023/go-sms-protocol/cmpp"
)

func TestCancel_RoundTripAndDispatcher(t *testing.T) {
	requestRaw, err := hex.DecodeString("0000001400000007000003010102030405060708")
	if err != nil {
		t.Fatal(err)
	}
	var request Cancel
	if err := request.IDecode(requestRaw); err != nil {
		t.Fatalf("Cancel.IDecode() error = %v", err)
	}
	if request.CommandID != cmpp.CommandCancel || request.SequenceID != 0x301 || request.MsgID != 0x0102030405060708 {
		t.Fatalf("decoded cancel = %#v", request)
	}
	encoded, err := request.IEncode()
	if err != nil {
		t.Fatalf("Cancel.IEncode() error = %v", err)
	}
	if !bytes.Equal(encoded, requestRaw) {
		t.Fatalf("Cancel.IEncode() = %x, want %x", encoded, requestRaw)
	}
	pdu, err := DecodeCMPP30(requestRaw)
	if err != nil {
		t.Fatalf("DecodeCMPP30(cancel) error = %v", err)
	}
	if _, ok := pdu.(*Cancel); !ok {
		t.Fatalf("DecodeCMPP30(cancel) type = %T", pdu)
	}
	response := request.GenEmptyResponse()
	if response.GetCommand() != cmpp.CommandCancelResp || response.GetSequenceID() != request.GetSequenceID() {
		t.Fatalf("Cancel.GenEmptyResponse() = %#v", response)
	}

	responseRaw, err := hex.DecodeString("00000010800000070000030200000001")
	if err != nil {
		t.Fatal(err)
	}
	decodedResponse := new(CancelResp)
	if err := decodedResponse.IDecode(responseRaw); err != nil {
		t.Fatalf("CancelResp.IDecode() error = %v", err)
	}
	if decodedResponse.Header.SequenceID != 0x302 || decodedResponse.SuccessID != 1 {
		t.Fatalf("decoded cancel response = %#v", decodedResponse)
	}
	encoded, err = decodedResponse.IEncode()
	if err != nil {
		t.Fatalf("CancelResp.IEncode() error = %v", err)
	}
	if !bytes.Equal(encoded, responseRaw) {
		t.Fatalf("CancelResp.IEncode() = %x, want %x", encoded, responseRaw)
	}
	pdu, err = DecodeCMPP30(responseRaw)
	if err != nil {
		t.Fatalf("DecodeCMPP30(cancel resp) error = %v", err)
	}
	if _, ok := pdu.(*CancelResp); !ok {
		t.Fatalf("DecodeCMPP30(cancel resp) type = %T", pdu)
	}
}
