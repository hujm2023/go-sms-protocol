package sgip12

import (
	"bytes"
	"testing"

	"github.com/bytedance/mockey"
	"github.com/stretchr/testify/assert"

	"github.com/hujm2023/go-sms-protocol/sgip"
)

func TestBind(t *testing.T) {
	raw := []byte{
		0x0, 0x0, 0x0, 0x3d, 0x0, 0x0, 0x0, 0x1,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x1, 0x74, 0x65, 0x73,
		0x74, 0x75, 0x73, 0x65, 0x72, 0x6e, 0x61, 0x6d,
		0x65, 0x0, 0x0, 0x0, 0x0, 0x74, 0x65, 0x73,
		0x74, 0x70, 0x77, 0x64, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0, 0x0,
	}

	// 实现接口级别的单测
	a := new(Bind)
	assert.Nil(t, a.IDecode(raw))
	assert.Equal(t, sgip.SP_SMG, a.Type)
	assert.Equal(t, "testusername", a.Name)
	assert.Equal(t, "testpwd", a.Password)
	assert.Equal(t, sgip.SGIP_BIND, a.GetCommand())

	mockey.Mock(sgip.Timestamp).Return(0).Build()
	bind := &Bind{
		Header: sgip.Header{
			CommandID: sgip.SGIP_BIND,
			Sequence:  [3]uint32{0, 0, 0},
		},
		Type:     sgip.SP_SMG,
		Name:     "testusername",
		Password: "testpwd",
		Reserved: "",
	}
	value, err := bind.IEncode()
	assert.Nil(t, err)
	assert.True(t, bytes.EqualFold(raw, value))
	pdu, err := DecodeSGIP12(value)
	assert.Nil(t, err)
	_, ok := pdu.(*Bind)
	assert.True(t, ok)

	bindResp := bind.GenEmptyResponse()
	assert.Equal(t, sgip.SGIP_BIND_REP, bindResp.GetCommand())
	assert.Nil(t, bindResp.GenEmptyResponse())

	t.Log(a.String())
}

func TestBindResp(t *testing.T) {
	raw := []byte{
		0x0, 0x0, 0x0, 0x1d,
		0x80, 0x0, 0x0, 0x1,
		0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0,
		0x0, 0x0, 0x0, 0x0,
		0x0,
	}

	// 实现接口级别的单测
	a := new(BindResp)
	assert.Nil(t, a.IDecode(raw))

	encoded, err := a.IEncode()
	assert.Nil(t, err)
	assert.True(t, bytes.EqualFold(raw, encoded))
	response := &BindResp{
		Header: sgip.Header{
			CommandID: sgip.SGIP_BIND_REP,
			Sequence:  [3]uint32{0, 0, 0},
		},
		Result:   sgip.STAT_OK,
		Reserved: "",
	}

	value, err := response.IEncode()
	assert.Nil(t, err)
	assert.True(t, bytes.EqualFold(raw, value))

	pdu, err := DecodeSGIP12(value)
	assert.Nil(t, err)
	_, ok := pdu.(*BindResp)
	assert.True(t, ok)
}
