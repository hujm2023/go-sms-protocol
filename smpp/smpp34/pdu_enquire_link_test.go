package smpp34

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/hujm2023/go-sms-protocol/smpp"
)

type EnquireLinkTestSuite struct {
	suite.Suite

	enquireLinkBytes []byte
}

func (e *EnquireLinkTestSuite) SetupTest() {
	e.enquireLinkBytes = []byte{0, 0, 0, 16, 0, 0, 0, 21, 0, 0, 0, 0, 0, 0, 0, 5}
}

func (e *EnquireLinkTestSuite) TestEnquireLink_IEncode() {
	enquireLink := EnquireLink{
		Header: smpp.Header{
			ID:       smpp.ENQUIRE_LINK,
			Status:   smpp.ESME_ROK,
			Sequence: 5,
			Length:   16,
		},
	}
	data, err := enquireLink.IEncode()
	e.Nil(err)
	e.Equal(e.enquireLinkBytes, data)
}

func (e *EnquireLinkTestSuite) TestEnquireLink_IDecode() {
	enquireLink := new(EnquireLink)
	e.Nil(enquireLink.IDecode(e.enquireLinkBytes))
	e.Equal(smpp.ENQUIRE_LINK, enquireLink.Header.ID)
	e.Equal(smpp.ESME_ROK, enquireLink.Header.Status)
}

func (e *EnquireLinkTestSuite) TestEnquireLink_GetSequenceID() {
	enquireLink := new(EnquireLink)
	e.Nil(enquireLink.IDecode(e.enquireLinkBytes))
	e.Equal(uint32(5), enquireLink.GetSequenceID())
}

func (e *EnquireLinkTestSuite) TestEnquireLink_SetSequenceID() {
	enquireLink := new(EnquireLink)
	e.Nil(enquireLink.IDecode(e.enquireLinkBytes))
	enquireLink.SetSequenceID(12345)
	e.Equal(uint32(12345), enquireLink.GetSequenceID())
}

func (e *EnquireLinkTestSuite) TestEnquireLink_GetCommand() {
	enquireLink := new(EnquireLink)
	e.Nil(enquireLink.IDecode(e.enquireLinkBytes))
	e.Equal(smpp.ENQUIRE_LINK, enquireLink.GetCommand())
}

func (e *EnquireLinkTestSuite) TestEnquireLink_GenEmptyResponse() {
	enquireLink := new(EnquireLink)
	e.Nil(enquireLink.IDecode(e.enquireLinkBytes))
	enquireLinkResp, ok := enquireLink.GenEmptyResponse().(*EnquireLinkResp)
	e.True(ok)
	e.Equal(smpp.ENQUIRE_LINK_RESP, enquireLinkResp.GetCommand())
	e.Equal(uint32(5), enquireLinkResp.GetSequenceID())
}

func TestEnquireLinbkSuite(t *testing.T) {
	suite.Run(t, new(EnquireLinkTestSuite))
}
