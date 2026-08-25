package smpp34

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hujm2023/go-sms-protocol/smpp"
)

func TestDecodeSMPP34UnbindResp(t *testing.T) {
	pdu, err := DecodeSMPP34(NewUnBindRespBytes(11))
	assert.NoError(t, err)

	response, ok := pdu.(*UnBindResp)
	if !assert.True(t, ok) {
		return
	}
	assert.Equal(t, smpp.UNBIND_RESP, response.GetCommand())
	assert.Equal(t, uint32(11), response.GetSequenceID())
}
