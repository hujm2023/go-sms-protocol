package cmpp20

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hujm2023/go-sms-protocol/cmpp"
)

func TestPduDelivery(t *testing.T) {
	dataExpected := []byte{
		0x00, 0x00, 0x00, 0x67, 0x00, 0x00, 0x00, 0x05,
		0x00, 0x00, 0x00, 0x05, 0xb5, 0x25, 0x62, 0x80, 0x00, 0x01, 0x00, 0x00, 0x39, 0x30, 0x30, 0x30,
		0x30, 0x31, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x31, 0x33,
		0x34, 0x31, 0x32, 0x33, 0x34, 0x30, 0x30, 0x30, 0x30, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x12, 0x54, 0x68, 0x69, 0x73, 0x20, 0x69, 0x73, 0x20, 0x61, 0x20, 0x74,
		0x65, 0x73, 0x74, 0x20, 0x4d, 0x4f, 0x2e, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	d := new(PduDeliver)
	assert.Nil(t, d.IDecode(dataExpected))

	assert.Equal(t, uint32(0x67), d.TotalLength)
	assert.Equal(t, cmpp.CommandDeliver, d.CommandID)
	assert.Equal(t, uint32(0x67), d.TotalLength)

	assert.Equal(t, uint64(13052947396898652160), d.MsgID)
	assert.Equal(t, "900001", d.DestID)
	assert.Equal(t, uint8(0), d.MsgFmt)
	assert.Equal(t, "13412340000", d.SrcTerminalID)
	assert.Equal(t, uint8(0), d.RegisteredDeliver)
	assert.Equal(t, uint8(18), d.MsgLength)
	assert.Equal(t, "This is a test MO.", d.MsgContent)

	encoded, err := d.IEncode()
	assert.Nil(t, err)
	assert.True(t, bytes.Equal(dataExpected, encoded))

	t.Log(d.String())
}

func TestDelivery(t *testing.T) {
	b := []byte{0, 0, 0, 145, 0, 0, 0, 5, 0, 0, 0, 0, 174, 92, 107, 192, 0, 11, 0, 1, 49, 48, 54, 57, 48, 53, 52, 57, 50, 50, 50, 50, 50, 51, 0, 0, 0, 0, 0, 0, 0, 116, 101, 115, 116, 0, 0, 0, 0, 0, 0, 0, 0, 0, 49, 56, 48, 50, 54, 57, 48, 49, 48, 50, 52, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 60, 174, 92, 107, 192, 0, 11, 0, 1, 68, 69, 76, 73, 86, 82, 68, 50, 50, 49, 48, 50, 56, 50, 51, 48, 54, 50, 50, 49, 48, 50, 56, 50, 51, 48, 54, 49, 56, 48, 50, 54, 57, 48, 49, 48, 50, 52, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0}
	d := new(PduDeliver)
	t.Log(d.IDecode(b))
	t.Log([]byte(d.MsgContent))
	t.Log(d.String())
}
