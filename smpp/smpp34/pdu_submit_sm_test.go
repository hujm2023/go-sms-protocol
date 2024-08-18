package smpp34

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/hujm2023/go-sms-protocol/smpp"
)

type SubmitSmTestSuite struct {
	suite.Suite
	serviceType      string
	sourceAddr       string
	destAddr         string
	registerDelivery uint8
	esmClass         uint8
	dataCoding       uint8

	content string

	valueBytes []byte
}

func (s *SubmitSmTestSuite) SetupTest() {
	s.serviceType = ""
	s.sourceAddr = "12345"
	s.destAddr = "18000001024"
	s.registerDelivery = 1
	s.esmClass = smpp.ESM_CLASS_MSGMODE_DATAGRAM
	s.dataCoding = smpp.ENCODING_IA5
	s.content = "Consequatur sit accusantium perferendis aut voluptatem. Sit perferendis voluptatem accusantium consequatur aut. Accusantium perferendis aut consequatur s"

	s.valueBytes = []byte{
		0, 0, 0, 218, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 40, 0, 1, 1, 49, 50, 51, 52, 53, 0, 1, 1, 49, 56, 48, 48, 48, 48, 48, 49, 48, 50, 52, 0, 1, 1, 1, 0, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 49, 53, 48, 48, 48, 82, 0, 1, 1, 1, 1, 153, 67, 111, 110, 115, 101, 113, 117, 97, 116, 117, 114, 32, 115, 105, 116, 32, 97, 99, 99, 117, 115, 97, 110, 116, 105, 117, 109, 32, 112, 101, 114, 102, 101, 114, 101, 110, 100, 105, 115, 32, 97, 117, 116, 32, 118, 111, 108, 117, 112, 116, 97, 116, 101, 109, 46, 32, 83, 105, 116, 32, 112, 101, 114, 102, 101, 114, 101, 110, 100, 105, 115, 32, 118, 111, 108, 117, 112, 116, 97, 116, 101, 109, 32, 97, 99, 99, 117, 115, 97, 110, 116, 105, 117, 109, 32, 99, 111, 110, 115, 101, 113, 117, 97, 116, 117, 114, 32, 97, 117, 116, 46, 32, 65, 99, 99, 117, 115, 97, 110, 116, 105, 117, 109, 32, 112, 101, 114, 102, 101, 114, 101, 110, 100, 105, 115, 32, 97, 117, 116, 32, 99, 111, 110, 115, 101, 113, 117, 97, 116, 117, 114, 32, 115,
	}
}

func (s *SubmitSmTestSuite) TearDownTest() {
}

func (s *SubmitSmTestSuite) TestSubmitSM_IDecode() {
	submit := new(SubmitSm)
	assert.Nil(s.T(), submit.IDecode(s.valueBytes))

	assert.Equal(s.T(), submit.Header.Sequence, uint32(40))
	assert.Equal(s.T(), submit.Header.ID, smpp.SUBMIT_SM)
	assert.Equal(s.T(), s.sourceAddr, submit.SourceAddr)
	assert.Equal(s.T(), s.destAddr, submit.DestinationAddr)
	assert.Equal(s.T(), s.registerDelivery, submit.RegisteredDelivery)
	assert.Equal(s.T(), s.content, string(submit.ShortMessage))
}

func (s *SubmitSmTestSuite) TestSubmitSM_IEncode() {
	content := []byte(s.content)
	b := SubmitSm{
		Header: smpp.Header{
			Length:   0,
			ID:       smpp.SUBMIT_SM,
			Status:   0,
			Sequence: 40,
		},
		ServiceType:          "",
		SourceAddrTon:        1,
		SourceAddrNpi:        1,
		SourceAddr:           s.sourceAddr,
		DestAddrTon:          1,
		DestAddrNpi:          1,
		DestinationAddr:      s.destAddr,
		ESMClass:             1,
		ProtocolID:           1,
		PriorityFlag:         1,
		ScheduleDeliveryTime: "",
		ValidityPeriod:       "000000000015000R",
		RegisteredDelivery:   s.registerDelivery,
		ReplaceIfPresentFlag: 1,
		DataCoding:           1,
		SmDefaultMsgID:       1,
		SmLength:             uint8(len(content)),
		ShortMessage:         content,
	}
	data, err := b.IEncode()
	assert.Nil(s.T(), err)

	if !bytes.Equal(data, s.valueBytes) {
		s.T().Log(data)
		s.T().Log(s.valueBytes)
		s.T().Fatal()
	}
}

func (s *SubmitSmTestSuite) TestSubmitSM_SetSequenceID() {
	submit := new(SubmitSm)
	assert.Nil(s.T(), submit.IDecode(s.valueBytes))

	submit.SetSequenceID(12345)
	assert.Equal(s.T(), submit.Header.Sequence, uint32(12345))
}

func (s *SubmitSmTestSuite) TestSubmitSM_GetSequenceID() {
	submit := new(SubmitSm)
	assert.Nil(s.T(), submit.IDecode(s.valueBytes))

	assert.Equal(s.T(), submit.GetSequenceID(), uint32(40))
}

func (s *SubmitSmTestSuite) TestSubmitSM_GetCommand() {
	submit := new(SubmitSm)
	assert.Nil(s.T(), submit.IDecode(s.valueBytes))

	assert.Equal(s.T(), submit.GetCommand(), smpp.SUBMIT_SM)
}

func (s *SubmitSmTestSuite) TestSubmitSM_GenEmptyResponse() {
	submit := new(SubmitSm)
	assert.Nil(s.T(), submit.IDecode(s.valueBytes))

	resp := submit.GenEmptyResponse()
	assert.Equal(s.T(), resp.GetCommand(), smpp.SUBMIT_SM_RESP)
	assert.Equal(s.T(), resp.GetSequenceID(), uint32(40))
}

func TestSubmitSm(t *testing.T) {
	suite.Run(t, new(SubmitSmTestSuite))
}

type SubmitSmRespTestSuite struct {
	suite.Suite

	MessageID string

	submitSmRespBytes []byte
}

func (s *SubmitSmRespTestSuite) SetupTest() {
	s.submitSmRespBytes = []byte{
		0, 0, 0, 53, 128, 0, 0, 4, 0, 0, 0, 0, 82, 33, 172, 56, 49, 48, 48, 57, 52, 54, 101, 52, 45, 53, 97, 56, 102, 45, 52, 56, 53, 100, 45, 56, 101, 54, 52, 45, 101, 100, 102, 57, 97, 97, 51, 55, 55, 97, 50, 50, 0,
	}
	s.MessageID = "100946e4-5a8f-485d-8e64-edf9aa377a22"
}

func (s *SubmitSmRespTestSuite) TestSubmitSmResp_IDecode() {
	submitSmResp := new(SubmitSmResp)
	assert.Nil(s.T(), submitSmResp.IDecode(s.submitSmRespBytes))

	assert.Equal(s.T(), submitSmResp.Header.Sequence, uint32(1377938488))
	assert.Equal(s.T(), submitSmResp.Header.ID, smpp.SUBMIT_SM_RESP)
	assert.Equal(s.T(), s.MessageID, submitSmResp.MessageID)
}

func (s *SubmitSmRespTestSuite) TestSubmitSmResp_IEncode() {
	b := SubmitSmResp{
		Header: smpp.Header{
			Length:   53,
			ID:       smpp.SUBMIT_SM_RESP,
			Status:   smpp.ESME_ROK,
			Sequence: 1377938488,
		},
		MessageID: s.MessageID,
	}
	data, err := b.IEncode()
	assert.Nil(s.T(), err)

	assert.Equal(s.T(), s.submitSmRespBytes, data)
}

func (s *SubmitSmRespTestSuite) TestSubmitSmResp_SetSequenceID() {
	submitSmResp := new(SubmitSmResp)
	assert.Nil(s.T(), submitSmResp.IDecode(s.submitSmRespBytes))

	submitSmResp.SetSequenceID(12345)
	assert.Equal(s.T(), submitSmResp.Header.Sequence, uint32(12345))
}

func (s *SubmitSmRespTestSuite) TestSubmitSmResp_GetSequenceID() {
	submitSmResp := new(SubmitSmResp)
	assert.Nil(s.T(), submitSmResp.IDecode(s.submitSmRespBytes))

	assert.Equal(s.T(), submitSmResp.GetSequenceID(), uint32(1377938488))
}

func (s *SubmitSmRespTestSuite) TestSubmitSmResp_GetCommand() {
	submitSmResp := new(SubmitSmResp)
	assert.Nil(s.T(), submitSmResp.IDecode(s.submitSmRespBytes))

	assert.Equal(s.T(), submitSmResp.GetCommand(), smpp.SUBMIT_SM_RESP)
}

func (s *SubmitSmRespTestSuite) TestSubmitSmResp_GenEmptyResponse() {
	submitSmResp := new(SubmitSmResp)
	assert.Nil(s.T(), submitSmResp.IDecode(s.submitSmRespBytes))

	resp := submitSmResp.GenEmptyResponse()
	assert.Nil(s.T(), resp)
}

func TestSubmitRespSuite(t *testing.T) {
	suite.Run(t, new(SubmitSmRespTestSuite))
}
