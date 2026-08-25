package sgip12

import (
	"encoding/binary"
	"errors"
	"reflect"
	"testing"

	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/sgip"
)

const maxSGIP12FuzzInput = 4096

const sgipDeliverMessageLengthOffset = sgip.HeaderLength + 21 + 21 + 1 + 1 + 1

// FuzzDecodeSGIP12 exercises command dispatch and the PDU encode/decode
// contract. The input's declared total length is intentionally not required to
// match the input: several SGIP PDU decoders accept trailing bytes or leave
// total-length validation to the command-specific implementation.
func FuzzDecodeSGIP12(f *testing.F) {
	for _, seed := range sgip12FuzzSeeds(f) {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > maxSGIP12FuzzInput {
			data = data[:maxSGIP12FuzzInput]
		}

		header, headerErr := sgip.PeekHeader(data)
		if len(data) < sgip.MinSGIPPduLength {
			pdu, err := DecodeSGIP12(data)
			if err == nil || pdu != nil {
				t.Fatalf("DecodeSGIP12(%x) = (%T, %v), want truncated-header error", data, pdu, err)
			}
			return
		}
		if headerErr != nil {
			t.Fatalf("PeekHeader(%x) error = %v after minimum length check", data, headerErr)
		}

		if !sgip12SupportedCommand(header.CommandID) {
			pdu, err := DecodeSGIP12(data)
			if pdu != nil || !errors.Is(err, sms.ErrUnsupportedPacket) {
				t.Fatalf("DecodeSGIP12 unknown command %#x = (%T, %v), want ErrUnsupportedPacket", header.CommandID, pdu, err)
			}
			return
		}

		pdu, err := DecodeSGIP12(data)
		if err != nil {
			if pdu != nil {
				t.Fatalf("DecodeSGIP12(%x) returned PDU %T with error: %v", data, pdu, err)
			}
			return
		}
		if pdu == nil {
			t.Fatalf("DecodeSGIP12(%x) returned nil PDU without error", data)
		}

		command := pdu.GetCommand()
		if command == nil {
			t.Fatalf("decoded %T returned nil command", pdu)
		}
		if got := command.ToUint32(); got != uint32(header.CommandID) {
			t.Fatalf("decoded %T command %#x, want %#x", pdu, got, header.CommandID)
		}
		if got := pdu.GetSequenceID(); got != header.Sequence[2] {
			t.Fatalf("decoded %T sequence %d, want %d", pdu, got, header.Sequence[2])
		}

		encoded, err := pdu.IEncode()
		if err != nil {
			t.Fatalf("re-encoding decoded %T: %v", pdu, err)
		}
		encodedHeader, err := sgip.PeekHeader(encoded)
		if err != nil {
			t.Fatalf("PeekHeader(re-encoded %T) error = %v", pdu, err)
		}
		if encodedHeader.TotalLength != uint32(len(encoded)) {
			t.Fatalf("re-encoded %T total length %d, want %d", pdu, encodedHeader.TotalLength, len(encoded))
		}
		if encodedHeader.CommandID != header.CommandID || encodedHeader.Sequence != header.Sequence {
			t.Fatalf(
				"re-encoded %T header = %#v, want command %#x sequence %#v",
				pdu,
				encodedHeader,
				header.CommandID,
				header.Sequence,
			)
		}

		decodedAgain, err := DecodeSGIP12(encoded)
		if err != nil {
			t.Fatalf("decoding canonical %T: %v", pdu, err)
		}
		if reflect.TypeOf(decodedAgain) != reflect.TypeOf(pdu) {
			t.Fatalf("decoded canonical type = %T, want %T", decodedAgain, pdu)
		}
		canonicalAgain, err := decodedAgain.IEncode()
		if err != nil {
			t.Fatalf("re-encoding canonical %T: %v", decodedAgain, err)
		}
		canonicalAgainHeader, err := sgip.PeekHeader(canonicalAgain)
		if err != nil {
			t.Fatalf("PeekHeader(second %T) error = %v", decodedAgain, err)
		}
		if len(canonicalAgain) != len(encoded) || canonicalAgainHeader != encodedHeader {
			t.Fatalf(
				"canonical %T header/length changed: first=%#v/%d second=%#v/%d",
				pdu,
				encodedHeader,
				len(encoded),
				canonicalAgainHeader,
				len(canonicalAgain),
			)
		}
	})
}

func sgip12FuzzSeeds(t *testing.F) [][]byte {
	encoders := []sms.Encoder{
		&Bind{
			Header:   sgip.Header{CommandID: sgip.SGIP_BIND, Sequence: [3]uint32{1, 2, 3}},
			Type:     sgip.SP_SMG,
			Name:     "client",
			Password: "secret",
		},
		&BindResp{
			Header: sgip.Header{CommandID: sgip.SGIP_BIND_REP, Sequence: [3]uint32{1, 2, 3}},
			Result: sgip.STAT_OK,
		},
		&Unbind{Header: sgip.Header{CommandID: sgip.SGIP_UNBIND, Sequence: [3]uint32{4, 5, 6}}},
		&UnbindResp{Header: sgip.Header{CommandID: sgip.SGIP_UNBIND_REP, Sequence: [3]uint32{4, 5, 6}}},
		&Submit{
			Header:         sgip.Header{CommandID: sgip.SGIP_SUBMIT, Sequence: [3]uint32{7, 8, 9}},
			SpNumber:       "10690090",
			ChargeNumber:   "8613800138000",
			UserCount:      1,
			UserNumber:     []string{"8613800138000"},
			CorpID:         "12345",
			ServiceType:    "svc",
			ExpireTime:     "yymmddhhmmsstnnp",
			ScheduleTime:   "yymmddhhmmsstnnp",
			MessageLength:  5,
			MessageContent: []byte("hello"),
			Reserved:       "",
		},
		&SubmitResp{Header: sgip.Header{CommandID: sgip.SGIP_SUBMIT_REP, Sequence: [3]uint32{7, 8, 9}}},
		&Report{
			Header:         sgip.Header{CommandID: sgip.SGIP_REPORT, Sequence: [3]uint32{10, 11, 12}},
			SubmitSequence: [3]uint32{13, 14, 15},
			UserNumber:     "8613800138000",
			State:          sgip.REPORT_OK,
			ErrorCode:      sgip.STAT_OK,
		},
		&ReportResp{Header: sgip.Header{CommandID: sgip.SGIP_REPORT_REP, Sequence: [3]uint32{10, 11, 12}}},
		&Deliver{
			Header:         sgip.Header{CommandID: sgip.SGIP_DELIVER, Sequence: [3]uint32{16, 17, 18}},
			UserNumber:     "8613800138000",
			SPNumber:       "10690090",
			MessageLength:  5,
			MessageContent: []byte("hello"),
			Reserved:       "",
		},
		&DeliverResp{Header: sgip.Header{CommandID: sgip.SGIP_DELIVER_REP, Sequence: [3]uint32{16, 17, 18}}},
	}

	seeds := make([][]byte, 0, len(encoders)+5)
	for _, encoder := range encoders {
		data, err := encoder.IEncode()
		if err != nil {
			t.Fatalf("building SGIP fuzz seed from %T: %v", encoder, err)
		}
		seeds = append(seeds, data)
	}

	seeds = append(seeds,
		nil,
		[]byte{},
		make([]byte, sgip.MinSGIPPduLength-1),
		sgip12HeaderSeed(sgip.MinSGIPPduLength, sgip.CommandID(0xffffffff)),
		sgip12HeaderSeed(1, sgip.SGIP_UNBIND),
		sgip12HeaderSeed(0xffffffff, sgip.SGIP_UNBIND_REP),
		sgip12OversizedDeliverSeed(),
	)
	return seeds
}

func sgip12HeaderSeed(totalLength uint32, command sgip.CommandID) []byte {
	data := make([]byte, sgip.HeaderLength)
	binary.BigEndian.PutUint32(data[0:4], totalLength)
	binary.BigEndian.PutUint32(data[4:8], uint32(command))
	binary.BigEndian.PutUint32(data[8:12], 1)
	binary.BigEndian.PutUint32(data[12:16], 2)
	binary.BigEndian.PutUint32(data[16:20], 3)
	return data
}

func sgip12SupportedCommand(command sgip.CommandID) bool {
	switch command {
	case sgip.SGIP_BIND,
		sgip.SGIP_BIND_REP,
		sgip.SGIP_UNBIND,
		sgip.SGIP_UNBIND_REP,
		sgip.SGIP_SUBMIT,
		sgip.SGIP_SUBMIT_REP,
		sgip.SGIP_REPORT,
		sgip.SGIP_REPORT_REP,
		sgip.SGIP_DELIVER,
		sgip.SGIP_DELIVER_REP:
		return true
	default:
		return false
	}
}

func sgip12OversizedDeliverSeed() []byte {
	data := make([]byte, sgipDeliverMessageLengthOffset+4)
	copy(data, sgip12HeaderSeed(uint32(len(data)), sgip.SGIP_DELIVER))
	binary.BigEndian.PutUint32(data[sgipDeliverMessageLengthOffset:], ^uint32(0))
	return data
}
