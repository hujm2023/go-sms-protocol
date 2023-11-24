package cmpp30

import (
	"testing"

	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ActiveTestTesstSuite struct {
	suite.Suite

	valueBytes []byte
}

func (s *ActiveTestTesstSuite) SetupTest() {
	s.valueBytes = []byte{
		0x00, 0x00, 0x00, 0x0c, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00, 0x17,
	}
}

func (s *ActiveTestTesstSuite) TestActiveTest_IEncode() {
	a := &ActiveTest{
		Header: cmpp.Header{
			CommandID:  cmpp.CommandActiveTest,
			SequenceID: 0x17,
		},
	}
	data, err := a.IEncode()
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), s.valueBytes, data)
}

func (s *ActiveTestTesstSuite) TestActiveTest_IDecode() {
	a := new(ActiveTest)
	assert.Nil(s.T(), a.IDecode(s.valueBytes))

	assert.Equal(s.T(), uint32(0x17), a.Header.SequenceID)
	assert.Equal(s.T(), cmpp.CommandActiveTest, a.Header.CommandID)
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
		0x00, 0x00, 0x00, 0x0d, 0x80, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00, 0x17, 0x00,
	}
}

func (s *ActiveTestRespTestSuite) TestActiveTestResp_IEncode() {
	a := &ActiveTestResp{
		Header: cmpp.Header{
			CommandID:  cmpp.CommandActiveTestResp,
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
	assert.Equal(s.T(), cmpp.CommandActiveTestResp, a.Header.CommandID)
}

func TestActiveTestResp(t *testing.T) {
	suite.Run(t, new(ActiveTestRespTestSuite))
}
