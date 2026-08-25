package smpp

import (
	"encoding/binary"
	"errors"
	"testing"
)

func TestValidateSequenceNumber(t *testing.T) {
	tests := []struct {
		name      string
		sequence  uint32
		allowZero bool
		wantErr   bool
	}{
		{name: "zero rejected", sequence: 0, wantErr: true},
		{name: "generic nack zero", sequence: 0, allowZero: true},
		{name: "minimum", sequence: SEQUENCE_NUM_START},
		{name: "maximum", sequence: SEQUENCE_NUM_END},
		{name: "sign bit rejected", sequence: 0x80000000, wantErr: true},
		{name: "uint32 maximum rejected", sequence: 0xffffffff, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSequenceNumber(tt.sequence, tt.allowZero)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateSequenceNumber(%#x, %t) error = %v, wantErr %t", tt.sequence, tt.allowZero, err, tt.wantErr)
			}
			if tt.wantErr && !errors.Is(err, ErrInvalidSequenceNumber) {
				t.Fatalf("error = %v, want ErrInvalidSequenceNumber", err)
			}
		})
	}
}

func TestValidateDecodedPDU(t *testing.T) {
	valid := make([]byte, MinSMPPPacketLen)
	binary.BigEndian.PutUint32(valid[0:4], MinSMPPPacketLen)
	binary.BigEndian.PutUint32(valid[4:8], uint32(SUBMIT_SM))
	binary.BigEndian.PutUint32(valid[12:16], 1)

	if _, err := ValidateDecodedPDU(valid, SUBMIT_SM, false); err != nil {
		t.Fatalf("valid PDU rejected: %v", err)
	}

	for _, tc := range []struct {
		name   string
		mutate func([]byte) []byte
	}{
		{name: "declared too small", mutate: func(p []byte) []byte { binary.BigEndian.PutUint32(p[:4], 15); return p }},
		{name: "declared differs", mutate: func(p []byte) []byte { binary.BigEndian.PutUint32(p[:4], 17); return p }},
		{name: "wrong command", mutate: func(p []byte) []byte { binary.BigEndian.PutUint32(p[4:8], uint32(DELIVER_SM)); return p }},
		{name: "invalid sequence", mutate: func(p []byte) []byte { binary.BigEndian.PutUint32(p[12:16], 0x80000000); return p }},
		{name: "trailing bytes", mutate: func(p []byte) []byte { return append(p, 0) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := append([]byte(nil), valid...)
			if _, err := ValidateDecodedPDU(tc.mutate(p), SUBMIT_SM, false); err == nil {
				t.Fatal("invalid PDU accepted")
			}
		})
	}
}

func TestValidateTimeCString(t *testing.T) {
	for _, value := range []string{"", "240825123045000+", "000000000015000R"} {
		if err := ValidateTimeCString("time", value); err != nil {
			t.Fatalf("valid time %q rejected: %v", value, err)
		}
	}

	for _, value := range []string{"24082512304500+", "241325123045000+", "000031000000000R", "000000000000001R"} {
		if err := ValidateTimeCString("time", value); err == nil {
			t.Fatalf("invalid time %q accepted", value)
		}
	}
}
