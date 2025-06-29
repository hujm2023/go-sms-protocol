package cmpp20

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hujm2023/go-sms-protocol/cmpp"
)

func TestPduTerminate(t *testing.T) {
	dataExpected := []byte{
		0x00, 0x00, 0x00, 0x0c, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x17,
	}
	tt := new(PduTerminate)
	assert.Nil(t, tt.IDecode(dataExpected))

	assert.Equal(t, uint32(0x0c), tt.TotalLength)
	assert.Equal(t, cmpp.CommandTerminate, tt.CommandID)
	assert.Equal(t, uint32(0x17), tt.SequenceID)

	data, err := tt.IEncode()
	assert.Nil(t, err)
	assert.Equal(t, dataExpected, data)
}
