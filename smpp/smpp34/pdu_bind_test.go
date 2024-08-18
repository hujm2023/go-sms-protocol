package smpp34

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

func (b *BindTestSuite) TestBind_IEncode() {
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

func (b *BindTestSuite) TestBind_IDecode() {
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

func (b *BindTestSuite) TestBind_GetSequenceID() {
	bind := new(Bind)
	assert.Nil(b.T(), bind.IDecode(b.bindBytes))

	assert.Equal(b.T(), uint32(1), bind.GetSequenceID())
}

func (b *BindTestSuite) TestBind_GetCommand() {
	bind := new(Bind)
	assert.Nil(b.T(), bind.IDecode(b.bindBytes))

	assert.Equal(b.T(), smpp.BIND_TRANSCEIVER, bind.GetCommand())
}

func (b *BindTestSuite) TestBind_GenEmptyResponse() {
	bind := new(Bind)
	assert.Nil(b.T(), bind.IDecode(b.bindBytes))

	bindResp, ok := bind.GenEmptyResponse().(*BindResp)
	assert.True(b.T(), ok)

	assert.Equal(b.T(), smpp.BIND_TRANSCEIVER_RESP, bindResp.GetCommand())
	assert.Equal(b.T(), uint32(1), bindResp.GetSequenceID())
}

func TestBind(t *testing.T) {
	suite.Run(t, new(BindTestSuite))
}

// ----------

type BindRespTestSuite struct {
	suite.Suite

	SystemID string
	tlvs     smpp.TLVs

	bindRespBytes []byte
}

func (b *BindRespTestSuite) SetupTest() {
	b.SystemID = "testing"
	b.tlvs.SetTLV(smpp.NewTLV(smpp.SC_INTERFACE_VERSION, []byte{52}))
	b.bindRespBytes = []byte{0, 0, 0, 29, 128, 0, 0, 9, 0, 0, 0, 0, 0, 0, 0, 4, 116, 101, 115, 116, 105, 110, 103, 0, 2, 16, 0, 1, 52}
}

func (b *BindRespTestSuite) TestBindResp_IEncode() {
	bindResp := BindResp{
		Header: smpp.Header{
			Length:   29,
			ID:       smpp.BIND_TRANSCEIVER_RESP,
			Status:   smpp.ESME_ROK,
			Sequence: 4,
		},
		SystemID: b.SystemID,
		TLVs:     b.tlvs,
	}

	data, err := bindResp.IEncode()
	assert.Nil(b.T(), err)

	assert.Equal(b.T(), b.bindRespBytes, data)
}

func (b *BindRespTestSuite) TestBindResp_IDecode() {
	bindResp := new(BindResp)
	assert.Nil(b.T(), bindResp.IDecode(b.bindRespBytes))

	assert.Equal(b.T(), b.SystemID, bindResp.SystemID)
	assert.Equal(b.T(), b.tlvs, bindResp.TLVs)
}

func (b *BindRespTestSuite) TestBindResp_SetSequenceID() {
	bindResp := new(BindResp)
	bindResp.SetSequenceID(123)

	assert.Equal(b.T(), uint32(123), bindResp.GetSequenceID())
}

func (b *BindRespTestSuite) TestBindResp_GetSequenceID() {
	bindResp := new(BindResp)
	assert.Nil(b.T(), bindResp.IDecode(b.bindRespBytes))

	assert.Equal(b.T(), uint32(4), bindResp.GetSequenceID())
}

func (b *BindRespTestSuite) TestBindResp_GetCommand() {
	bindResp := new(BindResp)
	assert.Nil(b.T(), bindResp.IDecode(b.bindRespBytes))

	assert.Equal(b.T(), smpp.BIND_TRANSCEIVER_RESP, bindResp.GetCommand())
}

func (b *BindRespTestSuite) TestBindResp_GenEmptyResponse() {
	bindResp := new(BindResp)
	assert.Nil(b.T(), bindResp.IDecode(b.bindRespBytes))

	assert.Nil(b.T(), bindResp.GenEmptyResponse())
}

func TestBindResp(t *testing.T) {
	suite.Run(t, new(BindRespTestSuite))
}
