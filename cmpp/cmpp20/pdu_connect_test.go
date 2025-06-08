package cmpp20

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hujm2023/go-sms-protocol/cmpp"
)

func TestPduConnect(t *testing.T) {
	dataExpected := []byte{
		0x00, 0x00, 0x00, 0x27, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x17, 0x39, 0x30, 0x30, 0x30,
		0x30, 0x31, 0x90, 0xd0, 0x0c, 0x1d, 0x51, 0x7a, 0xbd, 0x0b, 0x4f, 0x65, 0xf6, 0xbc, 0xf8, 0x53,
		0x5d, 0x16, 0x21, 0x3c, 0xdc, 0x73, 0xbe,
	}
	c := new(PduConnect)
	assert.Nil(t, c.IDecode(dataExpected))

	assert.Equal(t, uint32(0x27), c.TotalLength)
	assert.Equal(t, cmpp.CommandConnect, c.CommandID)
	assert.Equal(t, uint32(0x17), c.SequenceID)

	assert.Equal(t, "900001", c.SourceAddr)
	assert.Equal(t, uint8(0x21), c.Version)
	assert.Equal(t, uint32(1021080510), c.Timestamp)

	encoded, err := c.IEncode()
	assert.Nil(t, err)
	assert.True(t, bytes.Equal(encoded, dataExpected))

	t.Log(c.String())
}

func TestPduConnectResp(t *testing.T) {
	data := []byte{0, 0, 0, 30, 128, 0, 0, 1, 0, 0, 0, 38, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	c := new(PduConnectResp)
	assert.Nil(t, c.IDecode(data))
	assert.Equal(t, uint32(30), c.TotalLength)
	assert.Equal(t, cmpp.CommandConnectResp, c.CommandID)
	assert.Equal(t, uint32(38), c.SequenceID)
	assert.Equal(t, ConnectRespStatus(0), c.Status)
	assert.Equal(t, ConnectRespStatusSuccess.String(), c.Status.String())
	assert.Equal(t, "", c.AuthenticatorISMG)
	assert.Equal(t, uint8(0), c.Version)
	t.Log(c.String())
}
