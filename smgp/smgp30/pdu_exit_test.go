package smgp30

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hujm2023/go-sms-protocol/smgp"
)

func TestPduExit(t *testing.T) {
	dataExpected := []byte{
		0x0, 0x0, 0x0, 0xc, 0x0, 0x0, 0x0, 0x6, 0x0, 0x0, 0x4, 0xd2,
	}
	tt := new(Exit)
	assert.Nil(t, tt.IDecode(dataExpected))

	assert.Equal(t, uint32(0x0c), tt.TotalLength)
	assert.Equal(t, smgp.CommandExit, tt.CommandID)
	assert.Equal(t, uint32(1234), tt.SequenceID)
}
