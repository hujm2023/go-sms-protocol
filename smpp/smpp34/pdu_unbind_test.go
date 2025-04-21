package smpp34

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/hujm2023/go-sms-protocol/smpp"
)

type UnbindTestSuite struct {
	suite.Suite

	unbindBytes []byte
}

func (u *UnbindTestSuite) SetupTest() {
	u.unbindBytes = []byte{0, 0, 0, 16, 0, 0, 0, 6, 0, 0, 0, 0, 0, 0, 0, 3}
}

func (u *UnbindTestSuite) TestUnbind_IEncode() {
	unbind := Unbind{
		Header: smpp.Header{
			ID:       smpp.UNBIND,
			Status:   smpp.ESME_ROK,
			Sequence: 3,
			Length:   16,
		},
	}

	data, err := unbind.IEncode()
	assert.Nil(u.T(), err)
	assert.Equal(u.T(), u.unbindBytes, data)
}

func (u *UnbindTestSuite) TestUnbind_IDecode() {
	unbind := new(Unbind)
	assert.Nil(u.T(), unbind.IDecode(u.unbindBytes))
	assert.Equal(u.T(), unbind.Header.ID, smpp.UNBIND)
	assert.Equal(u.T(), unbind.Header.Status, smpp.ESME_ROK)
}

func (u *UnbindTestSuite) TestUnbind_SetSequenceID() {
	unbind := new(Unbind)
	assert.Nil(u.T(), unbind.IDecode(u.unbindBytes))
	assert.Equal(u.T(), unbind.GetSequenceID(), uint32(3))
	unbind.SetSequenceID(12345)
}

func (u *UnbindTestSuite) TestUnbind_GetSequenceID() {
	unbind := new(Unbind)
	assert.Nil(u.T(), unbind.IDecode(u.unbindBytes))
	assert.Equal(u.T(), unbind.GetSequenceID(), uint32(3))
}

func (u *UnbindTestSuite) TestUnbind_GetCommand() {
	unbind := new(Unbind)
	assert.Nil(u.T(), unbind.IDecode(u.unbindBytes))
	assert.Equal(u.T(), unbind.GetCommand(), smpp.UNBIND)
}

func (u *UnbindTestSuite) TestUnbind_GenEmptyResponse() {
	unbind := new(Unbind)
	assert.Nil(u.T(), unbind.IDecode(u.unbindBytes))
	resp, ok := unbind.GenEmptyResponse().(*UnBindResp)
	assert.True(u.T(), ok)
	assert.Equal(u.T(), resp.Header.ID, smpp.UNBIND_RESP)
	assert.Equal(u.T(), resp.Header.Sequence, uint32(3))
}

func TestUnbindSuite(t *testing.T) {
	suite.Run(t, new(UnbindTestSuite))
}

type UnbindRespTestSuite struct {
	suite.Suite

	unbindRespBytes []byte
}

func (u *UnbindRespTestSuite) SetupTest() {
	u.unbindRespBytes = []byte{0, 0, 0, 16, 128, 0, 0, 6, 0, 0, 0, 0, 0, 0, 0, 3}
}

func (u *UnbindRespTestSuite) TestUnbindResp_IEncode() {
	unbindResp := UnBindResp{
		Header: smpp.Header{
			ID:       smpp.UNBIND_RESP,
			Status:   smpp.ESME_ROK,
			Sequence: 3,
			Length:   16,
		},
	}

	data, err := unbindResp.IEncode()
	assert.Nil(u.T(), err)
	assert.Equal(u.T(), u.unbindRespBytes, data)
}

func (u *UnbindRespTestSuite) TestUnbindResp_IDecode() {
	unbindResp := new(UnBindResp)
	assert.Nil(u.T(), unbindResp.IDecode(u.unbindRespBytes))
	assert.Equal(u.T(), unbindResp.Header.ID, smpp.UNBIND_RESP)
	assert.Equal(u.T(), unbindResp.Header.Status, smpp.ESME_ROK)
}

func (u *UnbindRespTestSuite) TestUnbindResp_SetSequenceID() {
	unbindResp := new(UnBindResp)
	assert.Nil(u.T(), unbindResp.IDecode(u.unbindRespBytes))
	assert.Equal(u.T(), unbindResp.GetSequenceID(), uint32(3))
	unbindResp.SetSequenceID(12345)
	assert.Equal(u.T(), unbindResp.GetSequenceID(), uint32(12345))
}

func (u *UnbindRespTestSuite) TestUnbindResp_GetSequenceID() {
	unbindResp := new(UnBindResp)
	assert.Nil(u.T(), unbindResp.IDecode(u.unbindRespBytes))
	assert.Equal(u.T(), unbindResp.GetSequenceID(), uint32(3))
}

func (u *UnbindRespTestSuite) TestUnbindResp_GetCommand() {
	unbindResp := new(UnBindResp)
	assert.Nil(u.T(), unbindResp.IDecode(u.unbindRespBytes))
	assert.Equal(u.T(), unbindResp.GetCommand(), smpp.UNBIND_RESP)
}

func (u *UnbindRespTestSuite) TestUnbindResp_GenEmptyResponse() {
	unbindResp := new(UnBindResp)
	assert.Nil(u.T(), unbindResp.IDecode(u.unbindRespBytes))
	resp := unbindResp.GenEmptyResponse()
	assert.Nil(u.T(), resp)
}

func TestUnbindRespSuite(t *testing.T) {
	suite.Run(t, new(UnbindRespTestSuite))
}
