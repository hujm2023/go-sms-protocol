package smpp34

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"

	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/smpp"
)

func validSubmitSm() SubmitSm {
	message := []byte("hello")
	return SubmitSm{
		Header:          smpp.Header{ID: smpp.SUBMIT_SM, Sequence: 1},
		SourceAddrTon:   smpp.TON_International,
		SourceAddrNpi:   smpp.NPI_ISDN,
		SourceAddr:      "12345",
		DestAddrTon:     smpp.TON_International,
		DestAddrNpi:     smpp.NPI_ISDN,
		DestinationAddr: "18000001024",
		DataCoding:      smpp.ENCODING_DEFAULT,
		SmLength:        uint8(len(message)),
		ShortMessage:    message,
	}
}

func validDeliverSm() DeliverSm {
	message := []byte("hello")
	return DeliverSm{
		Header:          smpp.Header{ID: smpp.DELIVER_SM, Sequence: 1},
		SourceAddrTon:   smpp.TON_International,
		SourceAddrNpi:   smpp.NPI_ISDN,
		SourceAddr:      "18000001024",
		DestAddrTon:     smpp.TON_International,
		DestAddrNpi:     smpp.NPI_ISDN,
		DestinationAddr: "12345",
		DataCoding:      smpp.ENCODING_DEFAULT,
		SmLength:        uint8(len(message)),
		ShortMessage:    message,
	}
}

func validBind() Bind {
	return Bind{
		Header:           smpp.Header{ID: smpp.BIND_TRANSCEIVER, Sequence: 1},
		SystemID:         "system",
		Password:         "password",
		InterfaceVersion: 0x34,
		AddrTon:          smpp.TON_Unknown,
		AddrNpi:          smpp.NPI_Unknown,
	}
}

func TestSMPP34SequenceNumberDomainOnEncode(t *testing.T) {
	builders := []struct {
		name  string
		build func(uint32) sms.Encoder
	}{
		{name: "bind", build: func(seq uint32) sms.Encoder { p := validBind(); p.Sequence = seq; return &p }},
		{name: "bind response", build: func(seq uint32) sms.Encoder {
			return &BindResp{Header: smpp.Header{ID: smpp.BIND_TRANSCEIVER_RESP, Sequence: seq}}
		}},
		{name: "submit", build: func(seq uint32) sms.Encoder { p := validSubmitSm(); p.Sequence = seq; return &p }},
		{name: "submit response", build: func(seq uint32) sms.Encoder {
			return &SubmitSmResp{Header: smpp.Header{ID: smpp.SUBMIT_SM_RESP, Sequence: seq}}
		}},
		{name: "deliver", build: func(seq uint32) sms.Encoder { p := validDeliverSm(); p.Sequence = seq; return &p }},
		{name: "deliver response", build: func(seq uint32) sms.Encoder {
			return &DeliverSmResp{Header: smpp.Header{ID: smpp.DELIVER_SM_RESP, Sequence: seq}}
		}},
		{name: "enquire link", build: func(seq uint32) sms.Encoder {
			return &EnquireLink{Header: smpp.Header{ID: smpp.ENQUIRE_LINK, Sequence: seq}}
		}},
		{name: "enquire link response", build: func(seq uint32) sms.Encoder {
			return &EnquireLinkResp{Header: smpp.Header{ID: smpp.ENQUIRE_LINK_RESP, Sequence: seq}}
		}},
		{name: "unbind", build: func(seq uint32) sms.Encoder { return &Unbind{Header: smpp.Header{ID: smpp.UNBIND, Sequence: seq}} }},
		{name: "unbind response", build: func(seq uint32) sms.Encoder {
			return &UnBindResp{Header: smpp.Header{ID: smpp.UNBIND_RESP, Sequence: seq}}
		}},
	}

	for _, builder := range builders {
		for _, tc := range []struct {
			name    string
			seq     uint32
			wantErr bool
		}{
			{name: "zero", seq: 0, wantErr: true},
			{name: "minimum", seq: smpp.SEQUENCE_NUM_START},
			{name: "maximum", seq: smpp.SEQUENCE_NUM_END},
			{name: "sign bit", seq: 0x80000000, wantErr: true},
		} {
			t.Run(builder.name+"/"+tc.name, func(t *testing.T) {
				_, err := builder.build(tc.seq).IEncode()
				if (err != nil) != tc.wantErr {
					t.Fatalf("IEncode() error = %v, wantErr %t", err, tc.wantErr)
				}
			})
		}
	}

	for _, seq := range []uint32{0, smpp.SEQUENCE_NUM_END} {
		if _, err := (&GenericNack{Header: smpp.Header{ID: smpp.GENERIC_NACK, Status: smpp.ESME_RINVCMDLEN, Sequence: seq}}).IEncode(); err != nil {
			t.Fatalf("generic_nack sequence %#x rejected: %v", seq, err)
		}
	}
	if _, err := (&GenericNack{Header: smpp.Header{ID: smpp.GENERIC_NACK, Status: smpp.ESME_RINVCMDLEN, Sequence: 0x80000000}}).IEncode(); err == nil {
		t.Fatal("generic_nack accepted sequence_number above 0x7fffffff")
	}
}

func TestDecodeSMPP34ValidatesWholeHeader(t *testing.T) {
	p := validSubmitSm()
	data, err := p.IEncode()
	if err != nil {
		t.Fatal(err)
	}

	highSequence := append([]byte(nil), data...)
	binary.BigEndian.PutUint32(highSequence[12:16], 0x80000000)
	if _, err := DecodeSMPP34(highSequence); err == nil {
		t.Fatal("sequence_number above 0x7fffffff accepted")
	}

	trailing := append(append([]byte(nil), data...), 0)
	if _, err := DecodeSMPP34(trailing); err == nil {
		t.Fatal("bytes beyond command_length accepted")
	}

	truncatedTLV := append(append([]byte(nil), data...), 0, 1, 0)
	binary.BigEndian.PutUint32(truncatedTLV[:4], uint32(len(truncatedTLV)))
	if _, err := DecodeSMPP34(truncatedTLV); err == nil {
		t.Fatal("truncated TLV header accepted")
	}

	enquire, err := (&EnquireLink{Header: smpp.Header{ID: smpp.ENQUIRE_LINK, Sequence: 1}}).IEncode()
	if err != nil {
		t.Fatal(err)
	}
	enquire = append(enquire, 0)
	binary.BigEndian.PutUint32(enquire[:4], uint32(len(enquire)))
	if err := new(EnquireLink).IDecode(enquire); err == nil {
		t.Fatal("enquire_link body accepted")
	}
}

func TestSubmitSmProtocolFieldValidation(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*SubmitSm)
	}{
		{name: "wrong command id", mutate: func(p *SubmitSm) { p.ID = smpp.DELIVER_SM }},
		{name: "nonzero request status", mutate: func(p *SubmitSm) { p.Status = smpp.ESME_RSYSERR }},
		{name: "service type too long", mutate: func(p *SubmitSm) { p.ServiceType = "123456" }},
		{name: "source address too long", mutate: func(p *SubmitSm) { p.SourceAddr = "123456789012345678901" }},
		{name: "non ASCII address", mutate: func(p *SubmitSm) { p.SourceAddr = "短信" }},
		{name: "invalid TON", mutate: func(p *SubmitSm) { p.SourceAddrTon = 7 }},
		{name: "invalid NPI", mutate: func(p *SubmitSm) { p.SourceAddrNpi = 2 }},
		{name: "reserved esm class type", mutate: func(p *SubmitSm) { p.ESMClass = 0x04 }},
		{name: "invalid priority", mutate: func(p *SubmitSm) { p.PriorityFlag = 4 }},
		{name: "invalid schedule", mutate: func(p *SubmitSm) { p.ScheduleDeliveryTime = "000031000000000R" }},
		{name: "reserved registered delivery", mutate: func(p *SubmitSm) { p.RegisteredDelivery = 3 }},
		{name: "invalid replace flag", mutate: func(p *SubmitSm) { p.ReplaceIfPresentFlag = 2 }},
		{name: "reserved data coding", mutate: func(p *SubmitSm) { p.DataCoding = 0x0b }},
		{name: "reserved default message", mutate: func(p *SubmitSm) { p.SmDefaultMsgID = 0xff }},
		{name: "short message length mismatch", mutate: func(p *SubmitSm) { p.SmLength++ }},
		{name: "payload conflicts with short message", mutate: func(p *SubmitSm) { p.OptionalTLVs = smpp.TLVList{smpp.NewTLV(smpp.MESSAGE_PAYLOAD, []byte("payload"))} }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := validSubmitSm()
			tt.mutate(&p)
			if _, err := p.IEncode(); err == nil {
				t.Fatal("invalid submit_sm accepted")
			}
		})
	}

	for _, coding := range []uint8{0x00, 0x0a, 0x0d, 0x0e, 0xc0, 0xdf, 0xf0, 0xff} {
		p := validSubmitSm()
		p.DataCoding = coding
		if _, err := p.IEncode(); err != nil {
			t.Fatalf("defined data_coding %#02x rejected: %v", coding, err)
		}
	}
}

func TestDeliverSmProtocolFieldValidation(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*DeliverSm)
	}{
		{name: "reserved esm class type", mutate: func(p *DeliverSm) { p.ESMClass = 0x0c }},
		{name: "schedule must be null", mutate: func(p *DeliverSm) { p.ScheduleDeliveryTime = "000000000001000R" }},
		{name: "validity must be null", mutate: func(p *DeliverSm) { p.ValidityPeriod = "000000000001000R" }},
		{name: "registered delivery reserved bits", mutate: func(p *DeliverSm) { p.RegisteredDelivery = 1 }},
		{name: "replace flag must be zero", mutate: func(p *DeliverSm) { p.ReplaceIfPresentFlag = 1 }},
		{name: "default message must be zero", mutate: func(p *DeliverSm) { p.SmDefaultMsgId = 1 }},
		{name: "payload conflicts with short message", mutate: func(p *DeliverSm) {
			p.OptionalTLVs = smpp.TLVList{smpp.NewTLV(smpp.MESSAGE_PAYLOAD, []byte("payload"))}
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := validDeliverSm()
			tt.mutate(&p)
			if _, err := p.IEncode(); err == nil {
				t.Fatal("invalid deliver_sm accepted")
			}
		})
	}

	p := validDeliverSm()
	p.ESMClass = 0x04
	p.RegisteredDelivery = 0x0c
	if _, err := p.IEncode(); err != nil {
		t.Fatalf("defined deliver_sm flags rejected: %v", err)
	}
}

func TestOrderedOptionalTLVsRoundTripRepeatedTags(t *testing.T) {
	p := validSubmitSm()
	p.OptionalTLVs = smpp.TLVList{
		smpp.NewTLV(smpp.CALLBACK_NUM, []byte{1, 1, '1'}),
		smpp.NewTLV(smpp.CALLBACK_NUM, []byte{1, 1, '2'}),
	}
	data, err := p.IEncode()
	if err != nil {
		t.Fatal(err)
	}

	decoded := new(SubmitSm)
	if err := decoded.IDecode(data); err != nil {
		t.Fatal(err)
	}
	values := decoded.OptionalTLVs.All(smpp.CALLBACK_NUM)
	if len(values) != 2 || !bytes.Equal(values[0].Value(), []byte{1, 1, '1'}) || !bytes.Equal(values[1].Value(), []byte{1, 1, '2'}) {
		t.Fatalf("ordered repeated TLVs not retained: %#v", values)
	}
	if got := decoded.TLVs[smpp.CALLBACK_NUM].Value(); !bytes.Equal(got, []byte{1, 1, '2'}) {
		t.Fatalf("legacy TLV map should expose final occurrence, got %v", got)
	}
}

func TestMessagePayloadWithoutShortMessage(t *testing.T) {
	p := validSubmitSm()
	p.SmLength = 0
	p.ShortMessage = nil
	p.OptionalTLVs = smpp.TLVList{smpp.NewTLV(smpp.MESSAGE_PAYLOAD, []byte("payload"))}
	data, err := p.IEncode()
	if err != nil {
		t.Fatal(err)
	}

	decoded := new(SubmitSm)
	if err := decoded.IDecode(data); err != nil {
		t.Fatal(err)
	}
	if decoded.SmLength != 0 || len(decoded.ShortMessage) != 0 || len(decoded.OptionalTLVs.All(smpp.MESSAGE_PAYLOAD)) != 1 {
		t.Fatalf("message_payload representation changed: %#v", decoded)
	}

	p.OptionalTLVs = smpp.TLVList{smpp.NewTLV(smpp.MESSAGE_PAYLOAD, make([]byte, smpp.MAX_PDU_SIZE))}
	if _, err := p.IEncode(); err == nil {
		t.Fatal("PDU above the configured maximum size accepted")
	}
}

func TestResponseBodyRules(t *testing.T) {
	errorResponse := SubmitSmResp{
		Header:    smpp.Header{ID: smpp.SUBMIT_SM_RESP, Status: smpp.ESME_RSYSERR, Sequence: 1},
		MessageID: "ignored-on-error",
	}
	data, err := errorResponse.IEncode()
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != smpp.MinSMPPPacketLen {
		t.Fatalf("error submit_sm_resp length=%d, want 16", len(data))
	}
	if err := new(SubmitSmResp).IDecode(data); err != nil {
		t.Fatalf("header-only error response rejected: %v", err)
	}

	withBody := append(append([]byte(nil), data...), 0)
	binary.BigEndian.PutUint32(withBody[:4], uint32(len(withBody)))
	if err := new(SubmitSmResp).IDecode(withBody); err == nil {
		t.Fatal("error submit_sm_resp with body accepted")
	}

	deliverResponse := DeliverSmResp{Header: smpp.Header{ID: smpp.DELIVER_SM_RESP, Status: smpp.ESME_RSYSERR, Sequence: 1}}
	data, err = deliverResponse.IEncode()
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != smpp.MinSMPPPacketLen+1 || data[len(data)-1] != 0 {
		t.Fatalf("deliver_sm_resp must contain exactly one NULL body octet: %x", data)
	}
	headerOnly := append([]byte(nil), data[:smpp.MinSMPPPacketLen]...)
	binary.BigEndian.PutUint32(headerOnly[:4], smpp.MinSMPPPacketLen)
	if err := new(DeliverSmResp).IDecode(headerOnly); err == nil {
		t.Fatal("header-only deliver_sm_resp accepted")
	}
}

func TestBindFieldValidation(t *testing.T) {
	valid := validBind()
	if _, err := valid.IEncode(); err != nil {
		t.Fatal(err)
	}

	for _, mutate := range []func(*Bind){
		func(p *Bind) { p.SystemID = "1234567890123456" },
		func(p *Bind) { p.Password = "123456789" },
		func(p *Bind) { p.SystemType = "1234567890123" },
		func(p *Bind) { p.InterfaceVersion = 0x35 },
		func(p *Bind) { p.AddressRange = "12345678901234567890123456789012345678901" },
	} {
		p := validBind()
		mutate(&p)
		if _, err := p.IEncode(); err == nil {
			t.Fatal("invalid bind field accepted")
		}
	}
}

func TestCStringMaximumContentBoundaries(t *testing.T) {
	bind := validBind()
	bind.SystemID = strings.Repeat("s", 15)
	bind.Password = strings.Repeat("p", 8)
	bind.SystemType = strings.Repeat("t", 12)
	bind.AddressRange = strings.Repeat("a", 40)
	if _, err := bind.IEncode(); err != nil {
		t.Fatalf("maximum bind C-Octet String contents rejected: %v", err)
	}

	submit := validSubmitSm()
	submit.ServiceType = strings.Repeat("s", 5)
	submit.SourceAddr = strings.Repeat("1", 20)
	submit.DestinationAddr = strings.Repeat("2", 20)
	submit.ShortMessage = bytes.Repeat([]byte{'x'}, 254)
	submit.SmLength = 254
	if _, err := submit.IEncode(); err != nil {
		t.Fatalf("maximum submit_sm contents rejected: %v", err)
	}

	response := SubmitSmResp{
		Header:    smpp.Header{ID: smpp.SUBMIT_SM_RESP, Sequence: 1},
		MessageID: strings.Repeat("m", 64),
	}
	if _, err := response.IEncode(); err != nil {
		t.Fatalf("64-octet message_id rejected: %v", err)
	}
	response.MessageID += "m"
	if _, err := response.IEncode(); err == nil {
		t.Fatal("65-octet message_id content accepted")
	}
}
