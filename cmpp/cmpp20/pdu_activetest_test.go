package cmpp20

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPduActiveTest(t *testing.T) {
	dataExpected := []byte{
		0x00, 0x00, 0x00, 0x0c, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00, 0x17,
	}
	a := new(PduActiveTest)
	assert.Nil(t, a.IDecode(dataExpected))
	assert.Equal(t, uint32(12), a.TotalLength)
	assert.Equal(t, uint32(0x17), a.SequenceID)

	encoded, err := a.IEncode()
	assert.Nil(t, err)
	assert.True(t, bytes.EqualFold(dataExpected, encoded))

	t.Log(a.String())
}
