package smpp34

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/hujm2023/go-sms-protocol/smpp"
)

const maxSMPP34FuzzInput = smpp.MAX_PDU_SIZE

// FuzzDecodeSMPP34RoundTrip checks that accepted PDUs can be encoded into a
// canonical wire form and decoded again without changing their concrete type
// or serialized representation.
func FuzzDecodeSMPP34RoundTrip(f *testing.F) {
	for _, seed := range [][]byte{
		NewEnquireLinkReqBytes(1),
		NewEnquireLinkRespBytes(2),
		NewUnBindBytes(3),
		NewUnBindRespBytes(4),
		NewDeliverySMRespBytes(5),
	} {
		f.Add(seed)
	}

	for _, encoder := range []interface {
		IEncode() ([]byte, error)
	}{
		func() interface {
			IEncode() ([]byte, error)
		} {
			p := validSubmitSm()
			return &p
		}(),
		func() interface {
			IEncode() ([]byte, error)
		} {
			p := validDeliverSm()
			return &p
		}(),
		func() interface {
			IEncode() ([]byte, error)
		} {
			p := validBind()
			return &p
		}(),
		&GenericNack{Header: smpp.Header{ID: smpp.GENERIC_NACK, Status: smpp.ESME_RSYSERR, Sequence: 1}},
	} {
		seed, err := encoder.IEncode()
		if err != nil {
			f.Fatalf("building SMPP fuzz seed: %v", err)
		}
		f.Add(seed)
	}

	for _, seed := range [][]byte{
		nil,
		{},
		{0, 0, 0, 16},
		{0, 0, 0, 16, 0, 0, 0, 0xff, 0, 0, 0, 0, 0, 0, 0, 1},
		{0, 0, 0, 16, 0, 0, 0, 6, 0, 0, 0, 0, 0, 0, 0, 0},
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > maxSMPP34FuzzInput {
			data = data[:maxSMPP34FuzzInput]
		}

		pdu, err := DecodeSMPP34(data)
		if err != nil {
			if pdu != nil {
				t.Fatalf("invalid SMPP bytes returned PDU %T: %v", pdu, err)
			}
			return
		}

		canonical, err := pdu.IEncode()
		if err != nil {
			t.Fatalf("re-encoding decoded %T: %v", pdu, err)
		}
		if len(canonical) > smpp.MAX_PDU_SIZE {
			t.Fatalf("re-encoded %T PDU has length %d, exceeds %d", pdu, len(canonical), smpp.MAX_PDU_SIZE)
		}

		decoded, err := DecodeSMPP34(canonical)
		if err != nil {
			t.Fatalf("decoding canonical %T PDU: %v", pdu, err)
		}
		if reflect.TypeOf(decoded) != reflect.TypeOf(pdu) {
			t.Fatalf("decoded type = %T, want %T", decoded, pdu)
		}

		canonicalAgain, err := decoded.IEncode()
		if err != nil {
			t.Fatalf("re-encoding canonical %T PDU: %v", decoded, err)
		}
		if !bytes.Equal(canonicalAgain, canonical) {
			t.Fatalf("canonical encoding changed on second pass: first %x, second %x", canonical, canonicalAgain)
		}
	})
}
