package cmpp20

import (
	"bytes"
	"encoding/hex"
	"reflect"
	"testing"

	"github.com/hujm2023/go-sms-protocol/cmpp"
)

const maxCMPP20FuzzInput = 8 * 1024

type cmpp20FuzzEncoder interface {
	IEncode() ([]byte, error)
}

func decodeCMPP20FuzzSeed(value string) []byte {
	seed, err := hex.DecodeString(value)
	if err != nil {
		panic(err)
	}
	return seed
}

func addCMPP20EncodedSeed(f *testing.F, encoder cmpp20FuzzEncoder) {
	seed, err := encoder.IEncode()
	if err != nil {
		f.Fatalf("building CMPP 2.0 fuzz seed: %v", err)
	}
	f.Add(seed)
}

// FuzzDecodeCMPP20Dispatcher exercises command dispatch and canonical
// decode/encode stability for every CMPP 2.0 PDU type.
func FuzzDecodeCMPP20Dispatcher(f *testing.F) {
	for _, seed := range [][]byte{
		nil,
		{},
		{0, 0, 0, 12}, // incomplete header
		{0, 0, 0, 12, 0, 0, 0, 8, 0, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, 8, 0, 0, 0, 1}, // invalid declared length
		{0, 0, 0, 11, 0, 0, 0, 8, 0, 0, 0, 1},
		{0xff, 0xff, 0xff, 0xff, 0x80, 0, 0, 2, 0, 0, 0, 1}, // oversized declared length
		{0, 0, 0, 12, 0xde, 0xad, 0xbe, 0xef, 0, 0, 0, 1},   // unknown command
	} {
		f.Add(seed)
	}

	for _, seed := range []string{
		"0000000c0000000800000017",   // CMPP_ACTIVE_TEST
		"0000000d800000080000001700", // CMPP_ACTIVE_TEST_RESP
		"0000000c0000000200000017",   // CMPP_TERMINATE
		"0000000c8000000200000017",   // CMPP_TERMINATE_RESP
		"0000002700000001000000173930303030303190d00c1d517abd0b4f65f6bcf8535d16213cdc73be",                                               // CMPP_CONNECT
		"0000002700000006000001233230323630383235017376632d616c706861007273760000000000",                                                 // CMPP_QUERY
		"0000003f80000006000004563230323630383235017376632d616c706861000102030411121314212223243132333441424344515253546162636471727374", // CMPP_QUERY_RESP
	} {
		f.Add(decodeCMPP20FuzzSeed(seed))
	}

	addCMPP20EncodedSeed(f, &PduConnectResp{
		Header:  cmpp.NewHeader(0, cmpp.CommandConnectResp, 0x26),
		Status:  ConnectRespStatusSuccess,
		Version: 0x20,
	})
	addCMPP20EncodedSeed(f, &PduSubmit{
		Header:             cmpp.NewHeader(0, cmpp.CommandSubmit, 0x17),
		PkTotal:            1,
		PkNumber:           1,
		RegisteredDelivery: 1,
		MsgLevel:           1,
		ServiceID:          "test",
		FeeUserType:        2,
		FeeTerminalID:      "13500002696",
		MsgSrc:             "900001",
		FeeType:            "02",
		FeeCode:            "10",
		ValIDTime:          "151105131555101+",
		SrcID:              "900001",
		DestUsrTL:          1,
		DestTerminalID:     []string{"13500002696"},
		MsgLength:          17,
		MsgContent:         []byte("go submit content"),
	})
	addCMPP20EncodedSeed(f, &PduSubmitResp{
		Header: cmpp.NewHeader(0, cmpp.CommandSubmitResp, 0x17),
		MsgID:  0x0102030405060708,
		Result: 0,
	})
	addCMPP20EncodedSeed(f, &PduDeliver{
		Header:            cmpp.NewHeader(0, cmpp.CommandDeliver, 0x01),
		MsgID:             13052947396898652160,
		DestID:            "900001",
		SrcTerminalID:     "13412340000",
		RegisteredDeliver: 0,
		MsgLength:         18,
		MsgContent:        []byte("This is a test MO."),
	})
	addCMPP20EncodedSeed(f, &PduDeliverResp{
		Header: cmpp.NewHeader(0, cmpp.CommandDeliverResp, 0x01),
		MsgID:  0x0102030405060708,
		Result: 0,
	})

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > maxCMPP20FuzzInput {
			return
		}

		pdu, err := DecodeCMPP20(data)
		if err != nil {
			if pdu != nil {
				t.Fatalf("invalid CMPP 2.0 bytes returned PDU %T: %v", pdu, err)
			}
			return
		}
		if pdu == nil {
			t.Fatal("DecodeCMPP20 returned nil PDU without an error")
		}

		header, err := cmpp.PeekHeader(data)
		if err != nil {
			t.Fatalf("peeking accepted CMPP 2.0 header: %v", err)
		}
		if got := pdu.GetCommand().ToUint32(); got != header.CommandID.ToUint32() {
			t.Fatalf("decoded command = %#x, want %#x", got, header.CommandID.ToUint32())
		}
		if got := pdu.GetSequenceID(); got != header.SequenceID {
			t.Fatalf("decoded sequence = %d, want %d", got, header.SequenceID)
		}

		canonical, err := pdu.IEncode()
		if err != nil {
			t.Fatalf("re-encoding decoded %T: %v", pdu, err)
		}
		canonicalHeader, err := cmpp.PeekHeader(canonical)
		if err != nil {
			t.Fatalf("peeking canonical %T header: %v", pdu, err)
		}
		if got := canonicalHeader.TotalLength; got != uint32(len(canonical)) {
			t.Fatalf("canonical %T length field = %d, want %d", pdu, got, len(canonical))
		}
		if canonicalHeader.CommandID != header.CommandID || canonicalHeader.SequenceID != header.SequenceID {
			t.Fatalf("canonical %T header = %#v, want command %#x sequence %d", pdu, canonicalHeader, header.CommandID, header.SequenceID)
		}

		decoded, err := DecodeCMPP20(canonical)
		if err != nil {
			t.Fatalf("decoding canonical %T: %v", pdu, err)
		}
		if reflect.TypeOf(decoded) != reflect.TypeOf(pdu) {
			t.Fatalf("canonical decoded type = %T, want %T", decoded, pdu)
		}
		canonicalAgain, err := decoded.IEncode()
		if err != nil {
			t.Fatalf("re-encoding canonical %T: %v", decoded, err)
		}
		if !bytes.Equal(canonicalAgain, canonical) {
			t.Fatalf("canonical encoding changed for %T: first %x, second %x", pdu, canonical, canonicalAgain)
		}
	})
}
