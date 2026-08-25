package cmpp20

import (
	"encoding/binary"
	"errors"
	"testing"

	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/stretchr/testify/assert"
)

func TestNewConnect(t *testing.T) {
	c := NewConnect("900001", "888888", 0x17)
	data, err := c.IEncode()
	assert.Nil(t, err)
	assert.Equal(t, "900001", c.SourceAddr)
	assert.Equal(t, cmpp.CommandConnect, c.CommandID)
	assert.Equal(t, uint32(0x17), c.SequenceID)
	assert.Equal(t, uint8(cmpp.Version20), c.Version)
	assert.Len(t, data, MaxConnectLength)
	assert.Equal(t, uint32(len(data)), binary.BigEndian.Uint32(data[:cmpp.HeaderLength]))
}

func TestDecodeCMPP20_UnsupportedCommand(t *testing.T) {
	raw := []byte{0, 0, 0, 12, 0xde, 0xad, 0xbe, 0xef, 0, 0, 0, 1}
	if _, err := DecodeCMPP20(raw); !errors.Is(err, sms.ErrUnsupportedPacket) {
		t.Fatalf("DecodeCMPP20() error = %v, want %v", err, sms.ErrUnsupportedPacket)
	}
}
