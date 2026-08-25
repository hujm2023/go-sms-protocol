package smpp34

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/hujm2023/go-sms-protocol/smpp"
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

func TestBindCommandRouting(t *testing.T) {
	tests := []struct {
		name       string
		requestID  smpp.CMDId
		responseID smpp.CMDId
		sequenceID uint32
	}{
		{
			name:       "receiver",
			requestID:  smpp.BIND_RECEIVER,
			responseID: smpp.BIND_RECEIVER_RESP,
			sequenceID: 1,
		},
		{
			name:       "transmitter",
			requestID:  smpp.BIND_TRANSMITTER,
			responseID: smpp.BIND_TRANSMITTER_RESP,
			sequenceID: 2,
		},
		{
			name:       "transceiver",
			requestID:  smpp.BIND_TRANSCEIVER,
			responseID: smpp.BIND_TRANSCEIVER_RESP,
			sequenceID: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bind := &Bind{Header: smpp.Header{ID: tt.requestID, Sequence: tt.sequenceID}}
			assert.Equal(t, tt.requestID, bind.GetCommand())

			response, ok := bind.GenEmptyResponse().(*BindResp)
			if !assert.True(t, ok) {
				return
			}
			assert.Equal(t, tt.responseID, response.Header.ID)
			assert.Equal(t, tt.sequenceID, response.GetSequenceID())
			assert.Equal(t, tt.responseID, response.GetCommand())
		})
	}
}

func TestBindCommandRoutingUnknownID(t *testing.T) {
	bind := &Bind{Header: smpp.Header{ID: smpp.CMDId(0x7fffffff)}}
	assert.Nil(t, bind.GetCommand())
	assert.Nil(t, bind.GenEmptyResponse())

	bindResp := &BindResp{Header: smpp.Header{ID: smpp.CMDId(0x7fffffff)}}
	assert.Nil(t, bindResp.GetCommand())
}

func TestBindRespErrorHeaderOnly(t *testing.T) {
	bindResp := BindResp{
		Header: smpp.Header{
			ID:       smpp.BIND_RECEIVER_RESP,
			Status:   smpp.ESME_RINVPASWD,
			Sequence: 9,
		},
		SystemID: "must not be encoded",
	}
	bindResp.TLVs.SetTLV(smpp.NewTLV(smpp.SC_INTERFACE_VERSION, []byte{52}))

	data, err := bindResp.IEncode()
	assert.NoError(t, err)
	assert.Equal(t, []byte{
		0, 0, 0, 16,
		128, 0, 0, 1,
		0, 0, 0, 14,
		0, 0, 0, 9,
	}, data)

	decoded := new(BindResp)
	assert.NoError(t, decoded.IDecode(data))
	assert.Equal(t, smpp.BIND_RECEIVER_RESP, decoded.Header.ID)
	assert.Equal(t, smpp.ESME_RINVPASWD, decoded.Header.Status)
	assert.Equal(t, uint32(9), decoded.Header.Sequence)
	assert.Empty(t, decoded.SystemID)
	assert.Empty(t, decoded.TLVs)
}

func TestBindRespErrorRejectsTrailingBody(t *testing.T) {
	bindResp := BindResp{Header: smpp.Header{
		ID:       smpp.BIND_TRANSCEIVER_RESP,
		Status:   smpp.ESME_RBINDFAIL,
		Sequence: 10,
	}}
	data, err := bindResp.IEncode()
	assert.NoError(t, err)

	data = append(data, 0)
	assert.Error(t, new(BindResp).IDecode(data))
}
