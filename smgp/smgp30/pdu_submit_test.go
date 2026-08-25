package smgp30

import (
	"bytes"
	"reflect"
	"strings"
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

func TestSubmitIndexedDestTermIDFieldLengthError(t *testing.T) {
	p := &Submit{
		Header:          smgp.Header{CommandID: smgp.CommandSubmit, SequenceID: 1},
		DestTermIDCount: 1,
		DestTermID:      []string{strings.Repeat("d", 22)},
	}
	_, err := p.IEncode()
	requireFieldLengthError(t, err, "DestTermID[0]", 22, 21)
}

func TestSubmitEncodeRejectsInconsistentCountsAndLength(t *testing.T) {
	base := Submit{
		Header:          smgp.Header{CommandID: smgp.CommandSubmit, SequenceID: 1},
		DestTermIDCount: 1,
		DestTermID:      []string{"13800138000"},
		MsgLength:       1,
		MsgContent:      []byte{0x01},
	}

	tests := []struct {
		name   string
		mutate func(*Submit)
	}{
		{name: "destination count too small", mutate: func(p *Submit) { p.DestTermIDCount = 0 }},
		{name: "destination count too large", mutate: func(p *Submit) { p.DestTermIDCount = 2 }},
		{name: "message length too small", mutate: func(p *Submit) { p.MsgLength = 0 }},
		{name: "message length too large", mutate: func(p *Submit) { p.MsgLength = 2 }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := base
			tt.mutate(&p)
			if _, err := p.IEncode(); err == nil {
				t.Fatal("inconsistent Submit accepted")
			}
		})
	}
}

func TestSubmitBinaryContentAndOptionsRoundTrip(t *testing.T) {
	wantOptions := smgp.Options{}
	wantOptions.Add(smgp.NewOption(smgp.TAG_TP_pid, []byte{0x01, 0xff}))
	wantOptions.Add(smgp.NewOption(smgp.TAG_LinkID, []byte("link")))
	want := &Submit{
		Header:          smgp.Header{CommandID: smgp.CommandSubmit, SequenceID: 7},
		MsgType:         smgp.MT,
		MsgFormat:       smgp.BINARY,
		ServiceID:       "svc",
		SrcTermID:       "1069000000",
		DestTermIDCount: 1,
		DestTermID:      []string{"13800138000"},
		MsgLength:       4,
		MsgContent:      []byte{0x00, 0xff, 0x01, 0x80},
		Options:         wantOptions,
	}

	data, err := want.IEncode()
	if err != nil {
		t.Fatal(err)
	}
	got := new(Submit)
	if err := got.IDecode(data); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.MsgContent, want.MsgContent) {
		t.Fatalf("binary content=%x, want %x", got.MsgContent, want.MsgContent)
	}
	if !reflect.DeepEqual(got.Options, wantOptions) {
		t.Fatalf("options=%v, want %v", got.Options, wantOptions)
	}
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

func TestSMGP30MsgIDRequiresTenBytes(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{name: "empty", id: ""},
		{name: "odd hex", id: "0"},
		{name: "short", id: "010203040506070809"},
		{name: "long", id: "0102030405060708090102"},
		{name: "non hex", id: "0102030405060708090g"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &SubmitResp{
				Header: smgp.Header{CommandID: smgp.CommandSubmitResp, SequenceID: 1},
				MsgID:  tt.id,
			}
			if _, err := p.IEncode(); err == nil {
				t.Fatal("invalid MsgID accepted")
			}
		})
	}
}
