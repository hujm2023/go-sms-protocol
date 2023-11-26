package smpp50

import (
	"bytes"
	"testing"

	"github.com/hujm2023/go-sms-protocol/smpp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DeliverSmTestSuite struct {
	suite.Suite
	sourceAddr         string
	destAddr           string
	shortMessageString string
	msgID              string

	valueBytes []byte
}

func (s *DeliverSmTestSuite) SetupTest() {
	s.sourceAddr = "919800000285"
	s.destAddr = "SHAADI"
	s.shortMessageString = "id:7a44aaba-336f-4a92-9502-dd106aa7369f sub:001 dlvrd:001 submit date:231123193758 done date:231123193800 stat:DELIVRD err:000 text:"
	s.msgID = "7a44aaba-336f-4a92-9502-dd106aa7369f"

	s.valueBytes = []byte{
		0, 0, 0, 223, 0, 0, 0, 5, 0, 0, 0, 0, 0, 0, 98, 100, 0, 1, 1, 57, 49, 57, 56, 48, 48, 48, 48, 48, 50, 56, 53, 0, 1, 1, 83, 72, 65, 65, 68, 73, 0, 1, 1, 1, 0, 0, 1, 1, 1, 1, 132, 105, 100, 58, 55, 97, 52, 52, 97, 97, 98, 97, 45, 51, 51, 54, 102, 45, 52, 97, 57, 50, 45, 57, 53, 48, 50, 45, 100, 100, 49, 48, 54, 97, 97, 55, 51, 54, 57, 102, 32, 115, 117, 98, 58, 48, 48, 49, 32, 100, 108, 118, 114, 100, 58, 48, 48, 49, 32, 115, 117, 98, 109, 105, 116, 32, 100, 97, 116, 101, 58, 50, 51, 49, 49, 50, 51, 49, 57, 51, 55, 53, 56, 32, 100, 111, 110, 101, 32, 100, 97, 116, 101, 58, 50, 51, 49, 49, 50, 51, 49, 57, 51, 56, 48, 48, 32, 115, 116, 97, 116, 58, 68, 69, 76, 73, 86, 82, 68, 32, 101, 114, 114, 58, 48, 48, 48, 32, 116, 101, 120, 116, 58, 0, 30, 0, 36, 55, 97, 52, 52, 97, 97, 98, 97, 45, 51, 51, 54, 102, 45, 52, 97, 57, 50, 45, 57, 53, 48, 50, 45, 100, 100, 49, 48, 54, 97, 97, 55, 51, 54, 57, 102,
	}
}

func (s *DeliverSmTestSuite) TestDeliverSM_IDecode() {
}

func (s *DeliverSmTestSuite) TestDeliverSM_IEncode() {
	d := DeliverSm{
		Header: smpp.Header{
			ID:       smpp.DELIVER_SM,
			Sequence: 25188,
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
		ValidityPeriod:       "",
		RegisteredDelivery:   1,
		ReplaceIfPresentFlag: 1,
		DataCoding:           1,
		SmDefaultMsgID:       1,
		SmLength:             uint8(len([]byte(s.shortMessageString))),
		ShortMessage:         []byte(s.shortMessageString),
		tlv: map[uint16]smpp.TLV{
			smpp.RECEIPTED_MESSAGE_ID: smpp.NewTLVS(smpp.RECEIPTED_MESSAGE_ID, s.msgID),
		},
	}
	data, err := d.IEncode()
	assert.Nil(s.T(), err)
	assert.True(s.T(), bytes.Equal(data, s.valueBytes))
}

func (s *DeliverSmTestSuite) TestDeliverSM_SetSequenceID() {
	d := &DeliverSm{}
	assert.Nil(s.T(), d.IDecode(s.valueBytes))
	assert.Equal(s.T(), d.Header.ID, smpp.DELIVER_SM)
	assert.Equal(s.T(), d.Header.Sequence, uint32(25188))
	assert.Equal(s.T(), s.sourceAddr, d.SourceAddr)
	assert.Equal(s.T(), s.destAddr, d.DestinationAddr)
	assert.Equal(s.T(), s.shortMessageString, string(d.ShortMessage))

	tlv, ok := d.tlv[smpp.RECEIPTED_MESSAGE_ID]
	assert.True(s.T(), ok)
	assert.Equal(s.T(), s.msgID, string(tlv.Value))
}

func TestDeliverSmTestSuite(t *testing.T) {
	suite.Run(t, new(DeliverSmTestSuite))
}
