package smpp34

import (
	"encoding/binary"
	"errors"
	"reflect"
	"testing"

	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/smpp"
)

func TestDecodeSMPP34RoutesSupportedCommands(t *testing.T) {
	bindIDs := []smpp.CMDId{smpp.BIND_RECEIVER, smpp.BIND_TRANSMITTER, smpp.BIND_TRANSCEIVER}
	bindRespIDs := []smpp.CMDId{smpp.BIND_RECEIVER_RESP, smpp.BIND_TRANSMITTER_RESP, smpp.BIND_TRANSCEIVER_RESP}

	tests := []struct {
		name string
		pdu  sms.Encoder
		want interface{}
	}{
		{name: "enquire_link", pdu: &EnquireLink{Header: smpp.Header{ID: smpp.ENQUIRE_LINK, Sequence: 1}}, want: (*EnquireLink)(nil)},
		{name: "enquire_link_resp", pdu: &EnquireLinkResp{Header: smpp.Header{ID: smpp.ENQUIRE_LINK_RESP, Sequence: 1}}, want: (*EnquireLinkResp)(nil)},
		{name: "submit_sm", pdu: func() sms.Encoder { p := validSubmitSm(); return &p }(), want: (*SubmitSm)(nil)},
		{name: "submit_sm_resp", pdu: &SubmitSmResp{Header: smpp.Header{ID: smpp.SUBMIT_SM_RESP, Sequence: 1}}, want: (*SubmitSmResp)(nil)},
		{name: "deliver_sm", pdu: func() sms.Encoder { p := validDeliverSm(); return &p }(), want: (*DeliverSm)(nil)},
		{name: "deliver_sm_resp", pdu: &DeliverSmResp{Header: smpp.Header{ID: smpp.DELIVER_SM_RESP, Sequence: 1}}, want: (*DeliverSmResp)(nil)},
		{name: "unbind", pdu: &Unbind{Header: smpp.Header{ID: smpp.UNBIND, Sequence: 1}}, want: (*Unbind)(nil)},
		{name: "unbind_resp", pdu: &UnBindResp{Header: smpp.Header{ID: smpp.UNBIND_RESP, Sequence: 1}}, want: (*UnBindResp)(nil)},
		{name: "generic_nack", pdu: &GenericNack{Header: smpp.Header{ID: smpp.GENERIC_NACK, Status: smpp.ESME_RSYSERR, Sequence: 1}}, want: (*GenericNack)(nil)},
	}
	for _, id := range bindIDs {
		p := validBind()
		p.ID = id
		tests = append(tests, struct {
			name string
			pdu  sms.Encoder
			want interface{}
		}{name: id.String(), pdu: &p, want: (*Bind)(nil)})
	}
	for _, id := range bindRespIDs {
		tests = append(tests, struct {
			name string
			pdu  sms.Encoder
			want interface{}
		}{name: id.String(), pdu: &BindResp{Header: smpp.Header{ID: id, Sequence: 1}}, want: (*BindResp)(nil)})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.pdu.IEncode()
			if err != nil {
				t.Fatalf("IEncode() error = %v", err)
			}
			got, err := DecodeSMPP34(data)
			if err != nil {
				t.Fatalf("DecodeSMPP34() error = %v", err)
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Fatalf("DecodeSMPP34() type = %T, want %T", got, tt.want)
			}
		})
	}
}

func TestDecodeSMPP34RejectsUnknownCommand(t *testing.T) {
	data := make([]byte, smpp.MinSMPPPacketLen)
	binary.BigEndian.PutUint32(data[:4], uint32(len(data)))
	binary.BigEndian.PutUint32(data[4:8], uint32(0x000000ff))
	binary.BigEndian.PutUint32(data[8:12], uint32(smpp.ESME_ROK))
	binary.BigEndian.PutUint32(data[12:16], 1)

	_, err := DecodeSMPP34(data)
	if !errors.Is(err, sms.ErrUnsupportedPacket) {
		t.Fatalf("DecodeSMPP34() error = %v, want ErrUnsupportedPacket", err)
	}
}
