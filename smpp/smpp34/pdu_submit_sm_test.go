package smpp34

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
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
	s.destAddr = "18026901024"
	s.registerDelivery = 1
	s.esmClass = ESM_CLASS_MSGMODE_DATAGRAM
	s.dataCoding = ENCODING_IA5
	s.content = "Consequatur sit accusantium perferendis aut voluptatem. Sit perferendis voluptatem accusantium consequatur aut. Accusantium perferendis aut consequatur s"

	s.valueBytes = []byte{
		0, 0, 0, 219, 0, 0, 0, 4, 0, 0, 0, 0, 0, 0, 0, 40, 0, 1, 1, 49, 50, 51, 52, 53, 0, 1, 1, 49, 56, 48, 50, 54, 57, 48, 49, 48, 50, 52, 0, 0, 1, 1, 1, 0, 48, 48, 48, 48, 48, 48, 48, 48, 48, 48, 49, 53, 48, 48, 48, 82, 0, 1, 1, 1, 1, 153, 67, 111, 110, 115, 101, 113, 117, 97, 116, 117, 114, 32, 115, 105, 116, 32, 97, 99, 99, 117, 115, 97, 110, 116, 105, 117, 109, 32, 112, 101, 114, 102, 101, 114, 101, 110, 100, 105, 115, 32, 97, 117, 116, 32, 118, 111, 108, 117, 112, 116, 97, 116, 101, 109, 46, 32, 83, 105, 116, 32, 112, 101, 114, 102, 101, 114, 101, 110, 100, 105, 115, 32, 118, 111, 108, 117, 112, 116, 97, 116, 101, 109, 32, 97, 99, 99, 117, 115, 97, 110, 116, 105, 117, 109, 32, 99, 111, 110, 115, 101, 113, 117, 97, 116, 117, 114, 32, 97, 117, 116, 46, 32, 65, 99, 99, 117, 115, 97, 110, 116, 105, 117, 109, 32, 112, 101, 114, 102, 101, 114, 101, 110, 100, 105, 115, 32, 97, 117, 116, 32, 99, 111, 110, 115, 101, 113, 117, 97, 116, 117, 114, 32, 115,
	}
}

func (s *SubmitSmTestSuite) TearDownTest() {
}

func (s *SubmitSmTestSuite) TestSubmitSM_IDecode() {
}

func (s *SubmitSmTestSuite) TestSubmitSM_IEncode() {
	content := []byte(s.content)
	b := SubmitSm{
		Header: Header{
			Length:   0,
			ID:       SUBMIT_SM,
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
	}
}

func (s *SubmitSmTestSuite) TestSubmitSM_SetSequenceID() {
	submit := new(SubmitSm)
	assert.Nil(s.T(), submit.IDecode(s.valueBytes))

	assert.Equal(s.T(), submit.Header.Sequence, uint32(40))
	assert.Equal(s.T(), submit.Header.ID, SUBMIT_SM)
	assert.Equal(s.T(), s.sourceAddr, submit.SourceAddr)
	assert.Equal(s.T(), s.destAddr, submit.DestinationAddr)
	assert.Equal(s.T(), s.registerDelivery, submit.RegisteredDelivery)
	assert.Equal(s.T(), s.content, string(submit.ShortMessage))
}

func TestSubmitSm(t *testing.T) {
	suite.Run(t, new(SubmitSmTestSuite))
}
