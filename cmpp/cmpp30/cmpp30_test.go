package cmpp30

import (
	"errors"
	"testing"

	sms "github.com/hujm2023/go-sms-protocol"
)

func TestDecodeCMPP30_UnsupportedCommand(t *testing.T) {
	raw := []byte{0, 0, 0, 12, 0xde, 0xad, 0xbe, 0xef, 0, 0, 0, 1}
	if _, err := DecodeCMPP30(raw); !errors.Is(err, sms.ErrUnsupportedPacket) {
		t.Fatalf("DecodeCMPP30() error = %v, want %v", err, sms.ErrUnsupportedPacket)
	}
}
