package smpp50

import (
	"bytes"
	"testing"

	"github.com/hujm2023/go-sms-protocol/smpp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type BindTestSuite struct {
	suite.Suite

	systemID   string
	password   string
	version    uint8
	systemType string

	bindBytes []byte
}

func (b *BindTestSuite) SetupTest() {
	b.systemID = "h86g7v"
	b.password = "579024"
	b.version = uint8(0x34)
	b.systemType = "CMT"
	b.bindBytes = []byte{
		0, 0, 0, 38, 0, 0, 0, 9, 0, 0, 0, 0, 0, 0, 0, 1, 104, 56, 54, 103, 55, 118, 0, 53, 55, 57, 48, 50, 52, 0, 67, 77, 84, 0, 52, 0, 0, 0,
	}
}

func (b *BindTestSuite) TestBind_IDecode() {
	bind := Bind{
		Header: smpp.Header{
			Length:   0,
			ID:       smpp.BIND_TRANSCEIVER,
			Status:   0,
			Sequence: 1,
		},
		SystemID:         b.systemID,
		Password:         b.password,
		SystemType:       b.systemType,
		InterfaceVersion: b.version,
	}

	data, err := bind.IEncode()
	assert.Nil(b.T(), err)

	assert.True(b.T(), bytes.Equal(data, b.bindBytes))
}

func (b *BindTestSuite) TestBind_IEncode() {
	bind := new(Bind)
	assert.Nil(b.T(), bind.IDecode(b.bindBytes))

	assert.Equal(b.T(), b.systemID, bind.SystemID)
	assert.Equal(b.T(), b.password, bind.Password)
	assert.Equal(b.T(), b.systemType, bind.SystemType)
	assert.Equal(b.T(), b.version, bind.InterfaceVersion)
	assert.Equal(b.T(), uint8(0), bind.AddrTon)
	assert.Equal(b.T(), uint8(0), bind.AddrNpi)
	assert.Equal(b.T(), "", bind.AddressRange)
}

func (b *BindTestSuite) TestBind_SetSequenceID() {
	bind := new(Bind)
	bind.SetSequenceID(123)

	assert.Equal(b.T(), uint32(123), bind.Header.Sequence)
}

func TestBind(t *testing.T) {
	suite.Run(t, new(BindTestSuite))
}
