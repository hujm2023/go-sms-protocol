package smpp

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/hujm2023/go-sms-protocol/packet"
)

type TLVTestSuite struct {
	suite.Suite

	tag    uint16
	length uint16
	value  []byte

	valueBytes []byte
}

func (s *TLVTestSuite) SetupTest() {
	s.tag = RECEIPTED_MESSAGE_ID
	s.length = 6
	s.value = []byte("123456")
	s.valueBytes = []byte{
		0, 0x1e,
		0, 6,
		49, 50, 51, 52, 53, 54,
	}
}

func (s *TLVTestSuite) TestTLV_Bytes() {
	tt := NewTLV(s.tag, s.value)
	assert.True(s.T(), bytes.Equal(tt.Bytes(), s.valueBytes))
}

func (s *TLVTestSuite) TestReadTLVs() {
	r := packet.NewPacketReader(s.valueBytes)
	defer r.Release()

	t, err := ReadTLVs(r)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, len(t))
	tlv := t[s.tag]
	assert.Equal(s.T(), s.tag, tlv.Tag)
	assert.Equal(s.T(), s.length, tlv.Length)
	assert.True(s.T(), bytes.Equal(tlv.Value, s.value))
}

func TestTLV(t *testing.T) {
	suite.Run(t, new(TLVTestSuite))
}
