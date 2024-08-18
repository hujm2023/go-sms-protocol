package sgip12

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hujm2023/go-sms-protocol/sgip"
)

func TestUnBind(t *testing.T) {
	raw := []byte{
		0x0, 0x0, 0x0, 0x14,
		0x0, 0x0, 0x0, 0x2,
		0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0,
	}

	// 实现接口级别的单测
	a := new(Unbind)
	assert.Nil(t, a.IDecode(raw))
	unbind := Unbind{
		Header: sgip.Header{
			CommandID: sgip.SGIP_UNBIND,
			Sequence:  [3]uint32{0, 0, 0},
		},
	}
	value, err := unbind.IEncode()
	assert.Nil(t, err)
	assert.True(t, bytes.EqualFold(raw, value))

	pdu, err := DecodeSGIP12(value)
	assert.Nil(t, err)
	_, ok := pdu.(*Unbind)
	assert.True(t, ok)

	unbindResp := unbind.GenEmptyResponse()
	unbindResp.SetSequenceID(2)

	assert.Equal(t, sgip.SGIP_UNBIND_REP, unbindResp.GetCommand())
	assert.Nil(t, unbindResp.GenEmptyResponse())

	data, err := unbindResp.IEncode()
	assert.Nil(t, err)

	err = unbindResp.IDecode(data)
	assert.Nil(t, err)
}
