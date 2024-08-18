package smgp30

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/hujm2023/go-sms-protocol/smgp"
)

type ActiveTestTesstSuite struct {
	suite.Suite

	valueBytes []byte
}

func (s *ActiveTestTesstSuite) SetupTest() {
	s.valueBytes = []byte{
		0x0, 0x0, 0x0, 0xc, 0x0, 0x0, 0x0, 0x4, 0x0, 0x0, 0x4, 0xd2,
	}
}

func (s *ActiveTestTesstSuite) TestActiveTest_IEncode() {
	a := &ActiveTest{
		Header: smgp.Header{
			CommandID:   smgp.CommandActiveTest,
			SequenceID:  1234,
			TotalLength: smgp.HeaderLength,
		},
	}
	data, err := a.IEncode()
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), s.valueBytes, data)
}

func (s *ActiveTestTesstSuite) TestActiveTest_IDecode() {
	a := new(ActiveTest)
	assert.Nil(s.T(), a.IDecode(s.valueBytes))

	assert.Equal(s.T(), uint32(1234), a.Header.SequenceID)
	assert.Equal(s.T(), smgp.CommandActiveTest, a.Header.CommandID)
}

func TestActiveTest(t *testing.T) {
	suite.Run(t, new(ActiveTestTesstSuite))
}

type ActiveTestRespTestSuite struct {
	suite.Suite
	valueBytes []byte
}

func (s *ActiveTestRespTestSuite) SetupTest() {
	s.valueBytes = []byte{
		0x00, 0x00, 0x00, 0x0d, 0x80, 0x00, 0x00, 0x04, 0x00, 0x00, 0x00, 0x17, 0x00,
	}
}

func (s *ActiveTestRespTestSuite) TestActiveTestResp_IEncode() {
	a := &ActiveTestResp{
		Header: smgp.Header{
			CommandID:  smgp.CommandActiveTestResp,
			SequenceID: 0x17,
		},
	}
	data, err := a.IEncode()
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), s.valueBytes, data)
}

func (s *ActiveTestRespTestSuite) TestActiveTestResp_IDecode() {
	a := new(ActiveTestResp)
	assert.Nil(s.T(), a.IDecode(s.valueBytes))

	assert.Equal(s.T(), uint32(0x17), a.Header.SequenceID)
	assert.Equal(s.T(), smgp.CommandActiveTestResp, a.Header.CommandID)
}

func TestActiveTestResp(t *testing.T) {
	suite.Run(t, new(ActiveTestRespTestSuite))
}
