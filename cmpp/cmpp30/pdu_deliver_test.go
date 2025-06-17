package cmpp30

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/hujm2023/go-sms-protocol/cmpp"
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
		MsgContent:        []byte(s.msgContent),
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
	s.Equal([]byte(s.msgContent), deliver.MsgContent)
}

func (s *DeliverTestSuite) TestDeliver_SetSequenceID() {
	d := new(Deliver)
	s.Nil(d.IDecode(s.valueBytes))

	d.SetSequenceID(0x02)
	s.Equal(uint32(0x02), d.Header.SequenceID)
}

func (s *DeliverTestSuite) TestDeliver_GetSequenceID() {
	d := new(Deliver)
	s.Nil(d.IDecode(s.valueBytes))

	s.Equal(uint32(0x01), d.GetSequenceID())
}

func (s *DeliverTestSuite) TestDeliver_GetCommand() {
	d := new(Deliver)
	s.Nil(d.IDecode(s.valueBytes))

	s.Equal(cmpp.CommandDeliver, d.GetCommand())
}

func (s *DeliverTestSuite) TestDeliver_GenEmptyResponse() {
	d := new(Deliver)
	s.Nil(d.IDecode(s.valueBytes))

	resp := d.GenEmptyResponse()
	s.Equal(cmpp.CommandDeliverResp, resp.GetCommand())
	s.Equal(uint32(0x01), resp.GetSequenceID())
}

func (s *DeliverTestSuite) TestDeliver_String() {
	d := new(Deliver)
	s.Nil(d.IDecode(s.valueBytes))

	s.T().Log(d.String())
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

func (s *DeliverRespTestSuite) TestDeliverResp_SetSequenceID() {
	d := new(DeliverResp)
	s.Nil(d.IDecode(s.valueBytes))

	d.SetSequenceID(0x02)
	s.Equal(uint32(0x02), d.Header.SequenceID)
}

func (s *DeliverRespTestSuite) TestDeliverResp_GetSequenceID() {
	d := new(DeliverResp)
	s.Nil(d.IDecode(s.valueBytes))

	s.Equal(uint32(0x01), d.GetSequenceID())
}

func (s *DeliverRespTestSuite) TestDeliverResp_GetCommand() {
	d := new(DeliverResp)
	s.Nil(d.IDecode(s.valueBytes))

	s.Equal(cmpp.CommandDeliverResp, d.GetCommand())
}

func (s *DeliverRespTestSuite) TestDeliverResp_GenEmptyResponse() {
	d := new(DeliverResp)
	s.Nil(d.IDecode(s.valueBytes))

	resp := d.GenEmptyResponse()
	s.Nil(resp)
}

func TestDeliverResp(t *testing.T) {
	suite.Run(t, new(DeliverRespTestSuite))
}
