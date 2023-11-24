package cmpp30

import (
	"testing"

	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/stretchr/testify/suite"
)

type DeliverTestSuite struct {
	suite.Suite
	msgID            uint64
	destID           string
	msgFmt           uint8
	srcTerminalID    string
	registerDelivery uint8 // 0x01 means registered delivery, 0x00 is mo
	msgLength        uint8
	msgContent       string

	valueBytes []byte
}

func (s *DeliverTestSuite) SetupTest() {
	s.msgID = 13025908756704198656
	s.destID = "900001"
	s.msgFmt = 0
	s.srcTerminalID = "13412340000"
	s.registerDelivery = 0
	s.msgLength = 18
	s.msgContent = "This is a test MO."

	s.valueBytes = []byte{
		0x00, 0x00, 0x00, 0x7f, 0x00, 0x00, 0x00, 0x05,
		0x00, 0x00, 0x00, 0x01, 0xb4, 0xc5, 0x53, 0x00, 0x00, 0x01, 0x00, 0x00, 0x39, 0x30, 0x30, 0x30,
		0x30, 0x31, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x31, 0x33,
		0x34, 0x31, 0x32, 0x33, 0x34, 0x30, 0x30, 0x30, 0x30, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x12, 0x54, 0x68, 0x69, 0x73, 0x20, 0x69, 0x73, 0x20, 0x61, 0x20, 0x74, 0x65, 0x73, 0x74, 0x20,
		0x4d, 0x4f, 0x2e, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
}

func (s *DeliverTestSuite) TestDeliver_IEncode() {
	d := &Deliver{
		Header: cmpp.Header{
			CommandID:  cmpp.CommandDeliver,
			SequenceID: 0x01,
		},
		MsgID:             s.msgID,
		DestID:            s.destID,
		ServiceID:         "",
		MsgFmt:            s.msgFmt,
		SrcTerminalID:     s.srcTerminalID,
		SrcTerminalType:   0,
		RegisteredDeliver: s.registerDelivery,
		MsgLength:         s.msgLength,
		MsgContent:        s.msgContent,
	}

	data, err := d.IEncode()
	s.Nil(err)
	s.Equal(s.valueBytes, data)
}

func (s *DeliverTestSuite) TestDeliver_IDecode() {
	deliver := new(Deliver)
	s.Nil(deliver.IDecode(s.valueBytes))

	s.Equal(uint32(0x01), deliver.Header.SequenceID)
	s.Equal(cmpp.CommandDeliver, deliver.Header.CommandID)
	s.Equal(s.msgID, deliver.MsgID)
	s.Equal(s.destID, deliver.DestID)
	s.Equal(s.msgFmt, deliver.MsgFmt)
	s.Equal(s.srcTerminalID, deliver.SrcTerminalID)
	s.Equal(s.registerDelivery, deliver.RegisteredDeliver)
	s.Equal(s.msgLength, deliver.MsgLength)
	s.Equal(s.msgContent, deliver.MsgContent)
}

func TestDeliver(t *testing.T) {
	suite.Run(t, new(DeliverTestSuite))
}

type DeliverRespTestSuite struct {
	suite.Suite
	msgID      uint64
	valueBytes []byte
}

func (s *DeliverRespTestSuite) SetupTest() {
	s.msgID = 13025908756704198656
	s.valueBytes = []byte{
		0x00, 0x00, 0x00, 0x18, 0x80, 0x00, 0x00, 0x05, 0x00, 0x00, 0x00, 0x01, 0xb4, 0xc5, 0x53, 0x00,
		0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
}

func (s *DeliverRespTestSuite) TestDeliverResp_IEncode() {
	d := &DeliverResp{
		Header: cmpp.Header{
			CommandID:  cmpp.CommandDeliverResp,
			SequenceID: 0x01,
		},
		MsgID:  s.msgID,
		Result: 0,
	}
	data, err := d.IEncode()
	s.Nil(err)
	s.Equal(s.valueBytes, data)
}

func (s *DeliverRespTestSuite) TestDeliverResp_IDecode() {
	d := new(DeliverResp)
	s.Nil(d.IDecode(s.valueBytes))
	s.Equal(cmpp.CommandDeliverResp, d.Header.CommandID)
	s.Equal(uint32(0x01), d.Header.SequenceID)
	s.Equal(s.msgID, d.MsgID)
	s.Equal(uint32(0), d.Result)
}

func TestDeliverResp(t *testing.T) {
	suite.Run(t, new(DeliverRespTestSuite))
}
