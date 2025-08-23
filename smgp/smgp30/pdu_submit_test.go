package smgp30

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/hujm2023/go-sms-protocol/smgp"
)

type SubmitTestSuite struct {
	suite.Suite
	valueBytes []byte
}

func (s *SubmitTestSuite) SetupTest() {
	s.valueBytes = []byte{
		0x0, 0x0, 0x0, 0xa1, 0x0, 0x0, 0x0, 0x2, 0x0, 0x0, 0x4, 0xd2, 0x6, 0x1, 0x2, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x49, 0x64, 0x0, 0x30, 0x0, 0x30, 0x0, 0x0, 0x0, 0x0, 0x0, 0x30, 0x0, 0x0, 0x0, 0x0, 0x0, 0xf, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x31, 0x30, 0x36, 0x39, 0x30, 0x30, 0x30, 0x31, 0x31, 0x31, 0x31, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1, 0x31, 0x37, 0x36, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x30, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xe, 0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x74, 0x65, 0x73, 0x74, 0x20, 0x6d, 0x73, 0x67, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
	}
}

func (s *SubmitTestSuite) TestSubmit_IEncode() {
	submit := Submit{
		Header:          smgp.Header{TotalLength: 12, CommandID: smgp.CommandSubmit, SequenceID: 1234},
		MsgType:         smgp.MT,
		NeedReport:      smgp.NEED_REPORT,
		Priority:        smgp.HIGHER_PRIORITY,
		ServiceID:       "serviceId",
		FeeType:         "0",
		FeeCode:         "0",
		FixedFee:        "0",
		MsgFormat:       smgp.GB18030,
		ValidTime:       "",
		AtTime:          "",
		SrcTermID:       "10690001111",
		ChargeTermID:    "",
		DestTermIDCount: 1,
		DestTermID:      []string{"17600000000"},
		MsgLength:       uint8(len([]byte("hello test msg"))),
		MsgContent:      []byte("hello test msg"),
		Reserve:         "",
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
}

func (s *SubmitTestSuite) TestSubmit_IDecode() {
	submit := new(Submit)
	s.Nil(submit.IDecode(s.valueBytes))
	s.Equal(smgp.CommandSubmit, submit.Header.CommandID)
	s.Equal(uint32(1234), submit.Header.SequenceID)

	s.Equal(smgp.MT, submit.MsgType)
	s.Equal(smgp.NEED_REPORT, submit.NeedReport)
	s.Equal(smgp.HIGHER_PRIORITY, submit.Priority)
	s.Equal("serviceId", submit.ServiceID)
	s.Equal("0", submit.FeeType)
	s.Equal("0", submit.FeeCode)
	s.Equal("0", submit.FixedFee)
	s.Equal("", submit.ValidTime)
	s.Equal("10690001111", submit.SrcTermID)
	s.Equal(uint8(1), submit.DestTermIDCount)
	s.Equal([]string{"17600000000"}, submit.DestTermID)
	s.Equal(uint8(len([]byte("hello test msg"))), submit.MsgLength)
	s.Equal("hello test msg", string(submit.MsgContent))
}

func TestSubmit(t *testing.T) {
	suite.Run(t, new(SubmitTestSuite))
}

type SubmitRespTestSuite struct {
	suite.Suite

	valueBytes []byte
}

func (s *SubmitRespTestSuite) SetupTest() {
	s.valueBytes = []byte{
		0x0, 0x0, 0x0, 0x1a, 0x80, 0x0, 0x0, 0x2, 0x0, 0x0, 0x4, 0xd2, 0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0x1, 0x0, 0x0, 0x0, 0x0,
	}
}

func (s *SubmitRespTestSuite) TearDownTest() {
}

func (s *SubmitRespTestSuite) TestSubmitResp_IEncode() {
	submitResp := &SubmitResp{
		Header: smgp.Header{
			CommandID:  smgp.CommandSubmitResp,
			SequenceID: uint32(1234),
		},
		MsgID:  "01020304050607080901",
		Status: uint32(0),
	}
	data, err := submitResp.IEncode()
	s.Nil(err)
	s.Equal(s.valueBytes, data)
}

func (s *SubmitRespTestSuite) TestSubmitResp_IDecode() {
	submitResp := new(SubmitResp)
	s.Nil(submitResp.IDecode(s.valueBytes))
	s.Equal(smgp.CommandSubmitResp, submitResp.Header.CommandID)
	s.Equal(uint32(1234), submitResp.Header.SequenceID)
	s.Equal("01020304050607080901", submitResp.MsgID)
}

func TestSubmitResp(t *testing.T) {
	suite.Run(t, new(SubmitRespTestSuite))
}
