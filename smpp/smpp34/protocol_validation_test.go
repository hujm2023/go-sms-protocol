package smpp34

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

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

func assertGatewayRoundTrip(test *testing.T, original sms.PDU) []byte {
	test.Helper()
	wire, err := original.IEncode()
	require.NoError(test, err)
	require.NoError(test, original.IDecode(wire))
	directWire, err := original.IEncode()
	require.NoError(test, err)
	require.Equal(test, wire, directWire)
	decoded, err := DecodeSMPP34(wire)
	require.NoError(test, err)
	encoded, err := decoded.IEncode()
	require.NoError(test, err)
	require.Equal(test, wire, encoded)
	return wire
}

func TestGatewayCommandStatusCompatibility(test *testing.T) {
	for _, command := range []smpp.CMDId{
		smpp.BIND_RECEIVER, smpp.BIND_TRANSMITTER, smpp.BIND_TRANSCEIVER,
		smpp.SUBMIT_SM, smpp.DELIVER_SM, smpp.ENQUIRE_LINK, smpp.UNBIND, smpp.GENERIC_NACK,
	} {
		for _, status := range []uint32{0, uint32(smpp.ESME_RSYSERR), 0xffffffff} {
			test.Run(fmt.Sprintf("%s/status_%08x", command, status), func(subtest *testing.T) {
				var original sms.PDU
				switch command {
				case smpp.BIND_RECEIVER, smpp.BIND_TRANSMITTER, smpp.BIND_TRANSCEIVER:
					bind := validBind()
					bind.ID = command
					original = &bind
				case smpp.SUBMIT_SM:
					submit := validSubmitSm()
					original = &submit
				case smpp.DELIVER_SM:
					deliver := validDeliverSm()
					original = &deliver
				case smpp.ENQUIRE_LINK:
					original = &EnquireLink{Header: smpp.Header{ID: command, Sequence: 1}}
				case smpp.UNBIND:
					original = &Unbind{Header: smpp.Header{ID: command, Sequence: 1}}
				case smpp.GENERIC_NACK:
					original = &GenericNack{Header: smpp.Header{ID: command, Status: smpp.ESME_RSYSERR}}
				}
				wire, err := original.IEncode()
				require.NoError(subtest, err)
				binary.BigEndian.PutUint32(wire[8:12], status)
				require.NoError(subtest, original.IDecode(wire))
				require.Equal(subtest, wire, assertGatewayRoundTrip(subtest, original))
			})
		}
	}
}

func TestGatewayMessagePayloadCoexistence(test *testing.T) {
	for _, representation := range []string{"ordered list", "legacy map"} {
		test.Run(representation, func(subtest *testing.T) {
			submit := validSubmitSm()
			deliver := validDeliverSm()
			payload := []byte{0, 0xff, 'p'}
			tlvs := smpp.TLVList{smpp.NewTLV(smpp.MESSAGE_PAYLOAD, payload)}
			if representation == "ordered list" {
				submit.OptionalTLVs, deliver.OptionalTLVs = tlvs, tlvs
			} else {
				submit.TLVs, deliver.TLVs = tlvs.ToMap(), tlvs.ToMap()
			}
			assertGatewayRoundTrip(subtest, &submit)
			assertGatewayRoundTrip(subtest, &deliver)
			require.Equal(subtest, []byte("hello"), submit.ShortMessage)
			require.Equal(subtest, []byte("hello"), deliver.ShortMessage)
			require.Equal(subtest, uint8(5), submit.SmLength)
			require.Equal(subtest, uint8(5), deliver.SmLength)
			require.Equal(subtest, payload, submit.TLVs[smpp.MESSAGE_PAYLOAD].Value())
			require.Equal(subtest, payload, deliver.TLVs[smpp.MESSAGE_PAYLOAD].Value())
		})
	}
}

func TestGatewayResponseBodyCompatibility(test *testing.T) {
	for _, command := range []smpp.CMDId{
		smpp.BIND_RECEIVER_RESP, smpp.BIND_TRANSMITTER_RESP, smpp.BIND_TRANSCEIVER_RESP,
		smpp.SUBMIT_SM_RESP, smpp.DELIVER_SM_RESP,
	} {
		for _, status := range []uint32{0, uint32(smpp.ESME_RSYSERR)} {
			for _, body := range []struct {
				name string
				data []byte
			}{
				{name: "header only"},
				{name: "empty ID", data: []byte{0}},
				{name: "nonempty ID", data: []byte("gateway-id\x00")},
				{name: "unterminated ID", data: []byte("gateway-id")},
				{name: "trailing octet", data: []byte{0, 1}},
			} {
				test.Run(fmt.Sprintf("%s/status_%d/%s", command, status, body.name), func(subtest *testing.T) {
					var direct sms.PDU
					switch command {
					case smpp.SUBMIT_SM_RESP:
						direct = &SubmitSmResp{MessageID: "stale"}
					case smpp.DELIVER_SM_RESP:
						direct = &DeliverSmResp{MessageID: "stale"}
					default:
						direct = &BindResp{
							SystemID: "stale",
							TLVs:     smpp.TLVList{smpp.NewTLV(smpp.SC_INTERFACE_VERSION, []byte{0x34})}.ToMap(),
						}
					}
					wire := make([]byte, smpp.MinSMPPPacketLen)
					wire = append(wire, body.data...)
					binary.BigEndian.PutUint32(wire[:4], uint32(len(wire)))
					binary.BigEndian.PutUint32(wire[4:8], uint32(command))
					binary.BigEndian.PutUint32(wire[8:12], status)
					binary.BigEndian.PutUint32(wire[12:16], 1)
					decoded, err := DecodeSMPP34(wire)
					missingDeliverBody := command == smpp.DELIVER_SM_RESP && len(body.data) == 0
					malformedBody := body.name == "unterminated ID" || body.name == "trailing octet"
					if missingDeliverBody || malformedBody {
						require.Error(subtest, err)
						require.Error(subtest, direct.IDecode(wire))
						return
					}
					require.NoError(subtest, err)
					require.NoError(subtest, direct.IDecode(wire))
					for _, response := range []sms.PDU{decoded, direct} {
						encoded, err := response.IEncode()
						require.NoError(subtest, err)
						require.Equal(subtest, wire[4:16], encoded[4:16])
						expectedBody := body.data
						if len(body.data) <= 1 {
							expectedBody = []byte{0}
							if status != 0 && command != smpp.DELIVER_SM_RESP {
								expectedBody = []byte{}
							}
						}
						require.Equal(subtest, expectedBody, encoded[16:])
						_, err = DecodeSMPP34(encoded)
						require.NoError(subtest, err)
					}
				})
			}
		}
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
				if builder.name == "deliver" || builder.name == "deliver response" {
					tc.wantErr = false
				}
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
		{name: "service type contains NULL", mutate: func(p *SubmitSm) { p.ServiceType = "a\x00b" }},
		{name: "source address contains NULL", mutate: func(p *SubmitSm) { p.SourceAddr = "1\x002" }},
		{name: "NULL destination address", mutate: func(p *SubmitSm) { p.DestinationAddr = "1\x002" }},
		{name: "invalid TON", mutate: func(p *SubmitSm) { p.SourceAddrTon = 7 }},
		{name: "invalid NPI", mutate: func(p *SubmitSm) { p.SourceAddrNpi = 2 }},
		{name: "reserved esm class type", mutate: func(p *SubmitSm) { p.ESMClass = 0x04 }},
		{name: "invalid priority", mutate: func(p *SubmitSm) { p.PriorityFlag = 4 }},
		{name: "invalid schedule", mutate: func(p *SubmitSm) { p.ScheduleDeliveryTime = "000031000000000R" }},
		{name: "invalid replace flag", mutate: func(p *SubmitSm) { p.ReplaceIfPresentFlag = 2 }},
		{name: "reserved default message", mutate: func(p *SubmitSm) { p.SmDefaultMsgID = 0xff }},
		{name: "short message length mismatch", mutate: func(p *SubmitSm) { p.SmLength++ }},
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
		{name: "schedule contains NULL", mutate: func(p *DeliverSm) { p.ScheduleDeliveryTime = "a\x00b" }},
		{name: "validity contains NULL", mutate: func(p *DeliverSm) { p.ValidityPeriod = "a\x00b" }},
		{name: "replace flag must be zero", mutate: func(p *DeliverSm) { p.ReplaceIfPresentFlag = 1 }},
		{name: "default message must be zero", mutate: func(p *DeliverSm) { p.SmDefaultMsgId = 1 }},
		{name: "short message length mismatch", mutate: func(p *DeliverSm) { p.SmLength++ }},
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
		Header: smpp.Header{ID: smpp.SUBMIT_SM_RESP, Status: smpp.ESME_RSYSERR, Sequence: 1},
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
	if err := new(SubmitSmResp).IDecode(withBody); err != nil {
		t.Fatalf("error submit_sm_resp with NULL body rejected: %v", err)
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
		func(p *Bind) { p.SystemID = "system\x00id" },
		func(p *Bind) { p.Password = "pass\x00word" },
		func(p *Bind) { p.SystemID = "账号" },
		func(p *Bind) { p.Password = "密码" },
		func(p *Bind) { p.SystemType = "a\x00b" },
		func(p *Bind) { p.InterfaceVersion = 0x35 },
		func(p *Bind) { p.AddressRange = "1\x002" },
	} {
		p := validBind()
		mutate(&p)
		if _, err := p.IEncode(); err == nil {
			t.Fatal("invalid bind field accepted")
		}
	}
}

func TestStandardFieldLengthsRemainAccepted(t *testing.T) {
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
	if _, err := response.IEncode(); err != nil {
		t.Fatalf("extended message_id rejected: %v", err)
	}
}

func TestGatewayAddressBytesRoundTrip(t *testing.T) {
	for _, test := range []struct {
		name    string
		address string
	}{
		{name: "UTF-8", address: "短信签名"},
		{name: "non-UTF-8 octets", address: string([]byte{0xb6, 0xcc, 0xd0, 0xc5})},
	} {
		t.Run(test.name, func(t *testing.T) {
			submit := validSubmitSm()
			submit.SourceAddr, submit.DestinationAddr = test.address, test.address
			deliver := validDeliverSm()
			deliver.SourceAddr, deliver.DestinationAddr = test.address, test.address
			for _, original := range []sms.PDU{&submit, &deliver} {
				wire, err := original.IEncode()
				if err != nil {
					t.Fatal(err)
				}
				decoded, err := DecodeSMPP34(wire)
				if err != nil {
					t.Fatal(err)
				}
				encoded, err := decoded.IEncode()
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(wire, encoded) {
					t.Fatal("address bytes changed on round trip")
				}
			}
		})
	}
}

func TestExtendedVariableFieldsRoundTrip(t *testing.T) {
	bind := validBind()
	bind.SystemID = strings.Repeat("s", 128)
	bind.Password = strings.Repeat("p", 128)
	bind.SystemType = strings.Repeat("t", 128)
	bind.AddressRange = strings.Repeat("a", 128)
	submit := validSubmitSm()
	submit.ServiceType = strings.Repeat("s", 128)
	submit.SourceAddr = strings.Repeat("1", 128)
	submit.DestinationAddr = strings.Repeat("2", 128)
	submit.ShortMessage = bytes.Repeat([]byte{'x'}, 255)
	submit.SmLength = 255
	submit.OptionalTLVs = smpp.TLVList{
		smpp.NewTLV(smpp.RECEIPTED_MESSAGE_ID, append(bytes.Repeat([]byte{'m'}, 128), 0)),
	}
	deliver := validDeliverSm()
	deliver.ServiceType = submit.ServiceType
	deliver.SourceAddr = submit.SourceAddr
	deliver.DestinationAddr = submit.DestinationAddr
	deliver.ShortMessage = submit.ShortMessage
	deliver.SmLength = submit.SmLength
	deliver.OptionalTLVs = submit.OptionalTLVs
	response := &SubmitSmResp{
		Header:    smpp.Header{ID: smpp.SUBMIT_SM_RESP, Sequence: 1},
		MessageID: strings.Repeat("m", 128),
	}
	for _, test := range []struct {
		name     string
		original sms.PDU
		decoded  sms.PDU
	}{
		{name: "bind", original: &bind, decoded: new(Bind)},
		{name: "submit", original: &submit, decoded: new(SubmitSm)},
		{name: "deliver", original: &deliver, decoded: new(DeliverSm)},
		{name: "submit response", original: response, decoded: new(SubmitSmResp)},
	} {
		t.Run(test.name, func(t *testing.T) {
			wire, err := test.original.IEncode()
			if err != nil {
				t.Fatal(err)
			}
			if err := test.decoded.IDecode(wire); err != nil {
				t.Fatal(err)
			}
			encoded, err := test.decoded.IEncode()
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(wire, encoded) {
				t.Fatal("extended fields changed during round trip")
			}
		})
	}
	submit.ShortMessage = append(submit.ShortMessage, 'x')
	submit.SmLength = 0
	if _, err := submit.IEncode(); err == nil {
		t.Fatal("256-byte short_message encoded with uint8 length")
	}
	deliver.ShortMessage = submit.ShortMessage
	deliver.SmLength = 0
	if _, err := deliver.IEncode(); err == nil {
		t.Fatal("256-byte delivery encoded with uint8 length")
	}
}
