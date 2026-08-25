package sgip12

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/sgip"
)

func TestUnBind(t *testing.T) {
	raw := []byte{
		0x0, 0x0, 0x0, 0x14,
		0x0, 0x0, 0x0, 0x2,
		0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0,
	}

	// 实现接口级别的单测
	a := new(Unbind)
	assert.Nil(t, a.IDecode(raw))
	unbind := Unbind{
		Header: sgip.Header{
			CommandID: sgip.SGIP_UNBIND,
			Sequence:  [3]uint32{0, 0, 0},
		},
	}
	value, err := unbind.IEncode()
	assert.Nil(t, err)
	assert.True(t, bytes.EqualFold(raw, value))

	pdu, err := DecodeSGIP12(value)
	assert.Nil(t, err)
	_, ok := pdu.(*Unbind)
	assert.True(t, ok)

	unbindResp := unbind.GenEmptyResponse()
	unbindResp.SetSequenceID(2)

	assert.Equal(t, sgip.SGIP_UNBIND_REP, unbindResp.GetCommand())
	assert.Nil(t, unbindResp.GenEmptyResponse())

	data, err := unbindResp.IEncode()
	assert.Nil(t, err)

	err = unbindResp.IDecode(data)
	assert.Nil(t, err)
}

func TestUnbindResp_DispatchAndLengthErrors(t *testing.T) {
	raw := []byte{
		0x00, 0x00, 0x00, 0x14,
		0x80, 0x00, 0x00, 0x02,
		0x00, 0x00, 0x00, 0x0a,
		0x00, 0x00, 0x00, 0x14,
		0x00, 0x00, 0x00, 0x1e,
	}
	pdu, err := DecodeSGIP12(raw)
	if err != nil {
		t.Fatalf("DecodeSGIP12() error = %v", err)
	}
	response, ok := pdu.(*UnbindResp)
	if !ok {
		t.Fatalf("DecodeSGIP12() type = %T, want *UnbindResp", pdu)
	}
	if response.GetCommand() != sgip.SGIP_UNBIND_REP || response.Sequence != [3]uint32{10, 20, 30} {
		t.Fatalf("decoded response = %#v", response)
	}
	encoded, err := response.IEncode()
	if err != nil {
		t.Fatalf("IEncode() error = %v", err)
	}
	if !bytes.Equal(encoded, raw) {
		t.Fatalf("IEncode() = %x, want %x", encoded, raw)
	}

	tests := []struct {
		name string
		data []byte
	}{
		{name: "short header", data: raw[:sgip.MinSGIPPduLength-1]},
		{name: "declared length exceeds bytes", data: append([]byte{0x00, 0x00, 0x00, 0x1d}, raw[4:]...)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := DecodeSGIP12(tt.data); err == nil {
				t.Fatal("DecodeSGIP12() accepted malformed UnbindResp")
			}
		})
	}
}

func TestRequestResponses_PreserveSequence(t *testing.T) {
	sequence := [3]uint32{0x101, 0x202, 0x303}
	tests := []struct {
		name    string
		request sms.PDU
		command sgip.CommandID
		fullSeq func(sms.PDU) [3]uint32
	}{
		{
			name:    "bind",
			request: &Bind{Header: sgip.Header{CommandID: sgip.SGIP_BIND, Sequence: sequence}},
			command: sgip.SGIP_BIND_REP,
			fullSeq: func(pdu sms.PDU) [3]uint32 { return pdu.(*BindResp).Sequence },
		},
		{
			name:    "unbind",
			request: &Unbind{Header: sgip.Header{CommandID: sgip.SGIP_UNBIND, Sequence: sequence}},
			command: sgip.SGIP_UNBIND_REP,
			fullSeq: func(pdu sms.PDU) [3]uint32 { return pdu.(*UnbindResp).Sequence },
		},
		{
			name:    "deliver",
			request: &Deliver{Header: sgip.Header{CommandID: sgip.SGIP_DELIVER, Sequence: sequence}},
			command: sgip.SGIP_DELIVER_REP,
			fullSeq: func(pdu sms.PDU) [3]uint32 { return pdu.(*DeliverResp).Sequence },
		},
		{
			name:    "report",
			request: &Report{Header: sgip.Header{CommandID: sgip.SGIP_REPORT, Sequence: sequence}},
			command: sgip.SGIP_REPORT_REP,
			fullSeq: func(pdu sms.PDU) [3]uint32 { return pdu.(*ReportResp).Sequence },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := tt.request.GenEmptyResponse()
			if response == nil || response.GetCommand() != tt.command {
				t.Fatalf("GenEmptyResponse() = %#v, want command %v", response, tt.command)
			}
			if got := tt.fullSeq(response); got != sequence {
				t.Fatalf("response sequence = %#v, want %#v", got, sequence)
			}
		})
	}
}
