package smgp30

import (
	"encoding/binary"
	"errors"
	"reflect"
	"testing"

	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/smgp"
)

const maxSMGP30FuzzInput = 4096

const smgpDeliverMsgLengthOffset = smgp.HeaderLength + 10 + 1 + 1 + 14 + 21 + 21

// FuzzDecodeSMGP30 exercises command dispatch and the stable header/type
// contract of every supported SMGP 3.0 PDU. It deliberately checks headers
// after re-encoding instead of comparing whole wire payloads because Options
// are stored in a map and therefore have no wire order guarantee.
func FuzzDecodeSMGP30(f *testing.F) {
	for _, seed := range smgp30FuzzSeeds(f) {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > maxSMGP30FuzzInput {
			data = data[:maxSMGP30FuzzInput]
		}

		header, headerErr := smgp.PeekHeader(data)
		if len(data) < smgp.MinSMGPPduLength {
			pdu, err := DecodeSMGP30(data)
			if err == nil || pdu != nil {
				t.Fatalf("DecodeSMGP30(%x) = (%T, %v), want truncated-header error", data, pdu, err)
			}
			return
		}
		if headerErr != nil {
			t.Fatalf("PeekHeader(%x) error = %v after minimum length check", data, headerErr)
		}

		if !smgp30SupportedCommand(header.CommandID) {
			pdu, err := DecodeSMGP30(data)
			if pdu != nil || !errors.Is(err, sms.ErrUnsupportedPacket) {
				t.Fatalf("DecodeSMGP30 unknown command %#x = (%T, %v), want ErrUnsupportedPacket", header.CommandID, pdu, err)
			}
			return
		}

		pdu, err := DecodeSMGP30(data)
		if err != nil {
			if pdu != nil {
				t.Fatalf("DecodeSMGP30(%x) returned PDU %T with error: %v", data, pdu, err)
			}
			return
		}
		if pdu == nil {
			t.Fatalf("DecodeSMGP30(%x) returned nil PDU without error", data)
		}

		command := pdu.GetCommand()
		if command == nil {
			t.Fatalf("decoded %T returned nil command", pdu)
		}
		if got := command.ToUint32(); got != uint32(header.CommandID) {
			t.Fatalf("decoded %T command %#x, want %#x", pdu, got, header.CommandID)
		}
		if got := pdu.GetSequenceID(); got != header.SequenceID {
			t.Fatalf("decoded %T sequence %d, want %d", pdu, got, header.SequenceID)
		}

		encoded, err := pdu.IEncode()
		if err != nil {
			t.Fatalf("re-encoding decoded %T: %v", pdu, err)
		}
		encodedHeader, err := smgp.PeekHeader(encoded)
		if err != nil {
			t.Fatalf("PeekHeader(re-encoded %T) error = %v", pdu, err)
		}
		if encodedHeader.TotalLength != uint32(len(encoded)) {
			t.Fatalf("re-encoded %T total length %d, want %d", pdu, encodedHeader.TotalLength, len(encoded))
		}
		if encodedHeader.CommandID != header.CommandID || encodedHeader.SequenceID != header.SequenceID {
			t.Fatalf(
				"re-encoded %T header = %#v, want command %#x sequence %d",
				pdu,
				encodedHeader,
				header.CommandID,
				header.SequenceID,
			)
		}

		decodedAgain, err := DecodeSMGP30(encoded)
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
		canonicalAgainHeader, err := smgp.PeekHeader(canonicalAgain)
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

func smgp30FuzzSeeds(t *testing.F) [][]byte {
	options := smgp.Options{}
	options.Add(smgp.NewOption(smgp.TAG_TP_udhi, []byte{1}))

	encoders := []sms.Encoder{
		&Login{
			Header:              smgp.Header{CommandID: smgp.CommandLogin, SequenceID: 1},
			ClientID:            "client",
			AuthenticatorClient: string(make([]byte, 16)),
			LoginMode:           smgp.TRANSMIT_MODE,
			Version:             0x30,
			Timestamp:           101010101,
		},
		&LoginResp{
			Header:              smgp.Header{CommandID: smgp.CommandLoginResp, SequenceID: 2},
			Status:              LoginRespStatusSuccess,
			AuthenticatorServer: "server",
			ServerVersion:       0x30,
		},
		&Submit{
			Header:          smgp.Header{CommandID: smgp.CommandSubmit, SequenceID: 3},
			MsgType:         smgp.MT,
			NeedReport:      smgp.NEED_REPORT,
			Priority:        smgp.NORMAL_PRIORITY,
			ServiceID:       "svc",
			FeeType:         "0",
			FeeCode:         "0",
			FixedFee:        "0",
			MsgFormat:       smgp.GB18030,
			SrcTermID:       "1069000000",
			DestTermIDCount: 1,
			DestTermID:      []string{"13800138000"},
			MsgLength:       5,
			MsgContent:      []byte("hello"),
			Options:         options,
		},
		&SubmitResp{
			Header: smgp.Header{CommandID: smgp.CommandSubmitResp, SequenceID: 4},
			MsgID:  "01020304050607080901",
		},
		&Deliver{
			Header:     smgp.Header{CommandID: smgp.CommandDeliver, SequenceID: 5},
			MsgID:      "01020304050607080901",
			IsReport:   smgp.IS_REPORT,
			MsgFormat:  smgp.GB18030,
			SrcTermID:  "1069000000",
			DestTermID: "13800138000",
			MsgLength:  5,
			MsgContent: []byte("hello"),
			Options:    options,
		},
		&DeliverResp{
			Header: smgp.Header{CommandID: smgp.CommandDeliverResp, SequenceID: 6},
			MsgID:  "01020304050607080901",
		},
		&ActiveTest{Header: smgp.Header{CommandID: smgp.CommandActiveTest, SequenceID: 7}},
		&ActiveTestResp{Header: smgp.Header{CommandID: smgp.CommandActiveTestResp, SequenceID: 8}, Reserved: 1},
		&Exit{Header: smgp.Header{CommandID: smgp.CommandExit, SequenceID: 9}},
		&ExitResp{Header: smgp.Header{CommandID: smgp.CommandExitResp, SequenceID: 10}},
	}

	seeds := make([][]byte, 0, len(encoders)+5)
	for _, encoder := range encoders {
		data, err := encoder.IEncode()
		if err != nil {
			t.Fatalf("building SMGP fuzz seed from %T: %v", encoder, err)
		}
		seeds = append(seeds, data)
	}

	seeds = append(seeds,
		nil,
		[]byte{},
		make([]byte, smgp.MinSMGPPduLength-1),
		smgp30HeaderSeed(smgp.MinSMGPPduLength, smgp.CommandID(0xffffffff)),
		smgp30HeaderSeed(1, smgp.CommandExit),
		smgp30HeaderSeed(0xffffffff, smgp.CommandExitResp),
		smgp30OversizedDeliverOptionSeed(),
	)
	return seeds
}

func smgp30HeaderSeed(totalLength uint32, command smgp.CommandID) []byte {
	data := make([]byte, smgp.HeaderLength)
	binary.BigEndian.PutUint32(data[0:4], totalLength)
	binary.BigEndian.PutUint32(data[4:8], uint32(command))
	binary.BigEndian.PutUint32(data[8:12], 1)
	return data
}

func smgp30SupportedCommand(command smgp.CommandID) bool {
	switch command {
	case smgp.CommandLogin,
		smgp.CommandLoginResp,
		smgp.CommandSubmit,
		smgp.CommandSubmitResp,
		smgp.CommandDeliver,
		smgp.CommandDeliverResp,
		smgp.CommandActiveTest,
		smgp.CommandActiveTestResp,
		smgp.CommandExit,
		smgp.CommandExitResp:
		return true
	default:
		return false
	}
}

func smgp30OversizedDeliverOptionSeed() []byte {
	optionsOffset := smgpDeliverMsgLengthOffset + 1 + 8
	data := make([]byte, optionsOffset+4)
	copy(data, smgp30HeaderSeed(uint32(len(data)), smgp.CommandDeliver))
	binary.BigEndian.PutUint16(data[optionsOffset:optionsOffset+2], uint16(smgp.TAG_TP_pid))
	binary.BigEndian.PutUint16(data[optionsOffset+2:], ^uint16(0))
	return data
}
