package cmpp30

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/hujm2023/go-sms-protocol/cmpp"
)

var (
	pkTotal            uint8    = 1
	pkNumber           uint8    = 1
	registeredDelivery uint8    = 1
	msgLevel           uint8    = 1
	serviceId          string   = "test"
	feeUserType        uint8    = 2
	feeTerminalId      string   = "13500002696"
	feeTerminalType    uint8    = 0
	msgFmt             uint8    = 8
	msgSrc             string   = "900001"
	feeType            string   = "02"
	feeCode            string   = "10"
	validTime          string   = "151105131555101+"
	atTime             string   = ""
	srcId              string   = "900001"
	destUsrTl          uint8    = 1
	destTerminalId     []string = []string{"13500002696"}
	destTerminalType   uint8    = 0
	msgContent                  = "go submit"
	msgLength          uint8    = uint8(len(msgContent))
)

type SubmitTestSuite struct {
	suite.Suite
	valueBytes []byte
}

func (s *SubmitTestSuite) SetupTest() {
	s.valueBytes = []byte{
		0, 0, 0, 204, 0, 0, 0, 4, 0, 0, 0, 23, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 116, 101, 115, 116, 0, 0, 0, 0, 0, 0, 2, 49, 51, 53, 48, 48, 48, 48, 50, 54, 57, 54, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 8, 57, 48, 48, 48, 48, 49, 48, 50, 49, 48, 0, 0, 0, 0, 49, 53, 49, 49, 48, 53, 49, 51, 49, 53, 53, 53, 49, 48, 49, 43, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 57, 48, 48, 48, 48, 49, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 49, 51, 53, 48, 48, 48, 48, 50, 54, 57, 54, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 9, 103, 111, 32, 115, 117, 98, 109, 105, 116, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	}
}

func (s *SubmitTestSuite) TestSubmit_IEncode() {
	submit := &Submit{
		Header: cmpp.Header{
			CommandID:  cmpp.CommandSubmit,
			SequenceID: 0x17,
		},
		PkTotal:            pkTotal,
		PkNumber:           pkNumber,
		RegisteredDelivery: registeredDelivery,
		MsgLevel:           msgLevel,
		ServiceID:          serviceId,
		FeeUserType:        feeUserType,
		FeeTerminalID:      feeTerminalId,
		FeeTerminalType:    feeTerminalType,
		MsgFmt:             msgFmt,
		MsgSrc:             msgSrc,
		FeeType:            feeType,
		FeeCode:            feeCode,
		ValiDTime:          validTime,
		AtTime:             atTime,
		SrcID:              srcId,
		DestUsrTL:          destUsrTl,
		DestTerminalID:     destTerminalId,
		DestTerminalType:   destTerminalType,
		MsgLength:          msgLength,
		MsgContent:         msgContent,
	}
	data, err := submit.IEncode()
	s.Nil(err)
	s.Equal(s.valueBytes, data)
	for i := 0; i < len(s.valueBytes); i++ {
		if s.valueBytes[i] != data[i] {
			s.T().Log(data[:i+10])
			s.T().Log(s.valueBytes[:i+10])
			break
		}
	}
	s.T().Log(submit.String())
}

func (s *SubmitTestSuite) TestSubmit_IDecode() {
	submit := new(Submit)
	s.Nil(submit.IDecode(s.valueBytes))
	s.Equal(cmpp.CommandSubmit, submit.Header.CommandID)
	s.Equal(uint32(0x17), submit.Header.SequenceID)
	s.Equal(pkTotal, submit.PkTotal)
	s.Equal(pkNumber, submit.PkNumber)
	s.Equal(registeredDelivery, submit.RegisteredDelivery)
	s.Equal(msgLevel, submit.MsgLevel)
	s.Equal(serviceId, submit.ServiceID)
	s.Equal(feeUserType, submit.FeeUserType)
	s.Equal(feeTerminalId, submit.FeeTerminalID)
	s.Equal(feeTerminalType, submit.FeeTerminalType)
	s.Equal(msgFmt, submit.MsgFmt)
	s.Equal(msgSrc, submit.MsgSrc)
	s.Equal(feeType, submit.FeeType)
	s.Equal(feeCode, submit.FeeCode)
	s.Equal(validTime, submit.ValiDTime)
	s.Equal(atTime, submit.AtTime)
	s.Equal(srcId, submit.SrcID)
	s.Equal(destUsrTl, submit.DestUsrTL)
	s.Equal(destTerminalId, submit.DestTerminalID)
	s.Equal(destTerminalType, submit.DestTerminalType)
	s.Equal(msgLength, submit.MsgLength)
	s.Equal(msgContent, submit.MsgContent)
}

func (s *SubmitTestSuite) TestSubmit_SetSequenceID() {
	submit := new(Submit)
	s.Nil(submit.IDecode(s.valueBytes))

	id := uint32(0x20)
	submit.SetSequenceID(id)
	s.Equal(id, submit.GetSequenceID())
}

func (s *SubmitTestSuite) TestSubmit_GetSequenceID() {
	submit := new(Submit)
	s.Nil(submit.IDecode(s.valueBytes))

	s.Equal(uint32(0x17), submit.GetSequenceID())
}

func (s *SubmitTestSuite) TestSubmit_GetCommand() {
	submit := new(Submit)
	s.Nil(submit.IDecode(s.valueBytes))

	s.Equal(cmpp.CommandSubmit, submit.GetCommand())
}

func (s *SubmitTestSuite) TestSubmit_GenEmptyResponse() {
	submit := new(Submit)
	s.Nil(submit.IDecode(s.valueBytes))

	resp := submit.GenEmptyResponse()
	s.Equal(cmpp.CommandSubmitResp, resp.GetCommand())
	s.Equal(uint32(0x17), resp.GetSequenceID())
}

func TestSubmit(t *testing.T) {
	suite.Run(t, new(SubmitTestSuite))
}

type SubmitRespTestSuite struct {
	suite.Suite
	msgID uint64

	valueBytes []byte
}

func (s *SubmitRespTestSuite) SetupTest() {
	s.msgID = uint64(12878564852733378560)

	s.valueBytes = []byte{
		0x00, 0x00, 0x00, 0x18, 0x80, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x17,
		0xb2, 0xb9, 0xda, 0x80, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
}

func (s *SubmitRespTestSuite) TearDownTest() {
}

func (s *SubmitRespTestSuite) TestSubmitResp_IEncode() {
	submitResp := &SubmitResp{
		Header: cmpp.Header{
			CommandID:  cmpp.CommandSubmitResp,
			SequenceID: 0x17,
		},
		MsgID:  s.msgID,
		Result: 0,
	}
	data, err := submitResp.IEncode()
	s.Nil(err)
	s.Equal(s.valueBytes, data)
}

func (s *SubmitRespTestSuite) TestSubmitResp_IDecode() {
	submitResp := new(SubmitResp)
	s.Nil(submitResp.IDecode(s.valueBytes))
	s.Equal(cmpp.CommandSubmitResp, submitResp.Header.CommandID)
	s.Equal(uint32(0x17), submitResp.Header.SequenceID)
	s.Equal(s.msgID, submitResp.MsgID)
}

func (s *SubmitRespTestSuite) TestSubmitResp_SetSequenceID() {
	submitResp := new(SubmitResp)
	s.Nil(submitResp.IDecode(s.valueBytes))

	id := uint32(0x20)
	submitResp.SetSequenceID(id)
	s.Equal(id, submitResp.GetSequenceID())
}

func (s *SubmitRespTestSuite) TestSubmitResp_GetSequenceID() {
	submitResp := new(SubmitResp)
	s.Nil(submitResp.IDecode(s.valueBytes))

	s.Equal(uint32(0x17), submitResp.GetSequenceID())
}

func (s *SubmitRespTestSuite) TestSubmitResp_GetCommand() {
	submitResp := new(SubmitResp)
	s.Nil(submitResp.IDecode(s.valueBytes))

	s.Equal(cmpp.CommandSubmitResp, submitResp.GetCommand())
}

func (s *SubmitRespTestSuite) TestSubmitResp_GenEmptyResponse() {
	submitResp := new(SubmitResp)
	s.Nil(submitResp.IDecode(s.valueBytes))

	resp := submitResp.GenEmptyResponse()
	s.Nil(resp)
}

func TestSubmitResp(t *testing.T) {
	suite.Run(t, new(SubmitRespTestSuite))
}
