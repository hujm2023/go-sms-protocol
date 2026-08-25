package cmpp20

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/packet"
)

func TestPduConnect(t *testing.T) {
	dataExpected := []byte{
		0x00, 0x00, 0x00, 0x27, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x17, 0x39, 0x30, 0x30, 0x30,
		0x30, 0x31, 0x90, 0xd0, 0x0c, 0x1d, 0x51, 0x7a, 0xbd, 0x0b, 0x4f, 0x65, 0xf6, 0xbc, 0xf8, 0x53,
		0x5d, 0x16, 0x21, 0x3c, 0xdc, 0x73, 0xbe,
	}
	c := new(PduConnect)
	assert.Nil(t, c.IDecode(dataExpected))

	assert.Equal(t, uint32(0x27), c.TotalLength)
	assert.Equal(t, cmpp.CommandConnect, c.CommandID)
	assert.Equal(t, uint32(0x17), c.SequenceID)

	assert.Equal(t, "900001", c.SourceAddr)
	assert.Equal(t, uint8(0x21), c.Version)
	assert.Equal(t, uint32(1021080510), c.Timestamp)

	encoded, err := c.IEncode()
	assert.Nil(t, err)
	assert.True(t, bytes.Equal(encoded, dataExpected))

	t.Log(c.String())
}

func TestPduConnectResp(t *testing.T) {
	data := []byte{0, 0, 0, 30, 128, 0, 0, 1, 0, 0, 0, 38, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	c := new(PduConnectResp)
	assert.Nil(t, c.IDecode(data))
	assert.Equal(t, uint32(30), c.TotalLength)
	assert.Equal(t, cmpp.CommandConnectResp, c.CommandID)
	assert.Equal(t, uint32(38), c.SequenceID)
	assert.Equal(t, ConnectRespStatus(0), c.Status)
	assert.Equal(t, ConnectRespStatusSuccess.String(), c.Status.String())
	assert.Equal(t, string(make([]byte, 16)), c.AuthenticatorISMG)
	assert.Equal(t, uint8(0), c.Version)
	t.Log(c.String())
}

func TestPduConnectBinaryAuthenticatorsRoundTrip(t *testing.T) {
	authenticatorSource := string([]byte{
		0x01, 0x00, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
	})
	request := &PduConnect{
		Header:              cmpp.NewHeader(0, cmpp.CommandConnect, 1),
		SourceAddr:          "900001",
		AuthenticatorSource: authenticatorSource,
		Version:             cmpp.Version20,
		Timestamp:           825120000,
	}

	data, err := request.IEncode()
	if err != nil {
		t.Fatalf("PduConnect.IEncode() error = %v", err)
	}
	decodedRequest := new(PduConnect)
	if err := decodedRequest.IDecode(data); err != nil {
		t.Fatalf("PduConnect.IDecode() error = %v", err)
	}
	if decodedRequest.AuthenticatorSource != authenticatorSource {
		t.Fatalf("AuthenticatorSource = %x, want %x", decodedRequest.AuthenticatorSource, authenticatorSource)
	}

	authenticatorISMG := string([]byte{
		0x10, 0x0f, 0x0e, 0x0d, 0x0c, 0x00, 0x0a, 0x09,
		0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01,
	})
	response := &PduConnectResp{
		Header:            cmpp.NewHeader(0, cmpp.CommandConnectResp, 1),
		Status:            ConnectRespStatusSuccess,
		AuthenticatorISMG: authenticatorISMG,
		Version:           cmpp.Version20,
	}

	data, err = response.IEncode()
	if err != nil {
		t.Fatalf("PduConnectResp.IEncode() error = %v", err)
	}
	decodedResponse := new(PduConnectResp)
	if err := decodedResponse.IDecode(data); err != nil {
		t.Fatalf("PduConnectResp.IDecode() error = %v", err)
	}
	if decodedResponse.AuthenticatorISMG != authenticatorISMG {
		t.Fatalf("AuthenticatorISMG = %x, want %x", decodedResponse.AuthenticatorISMG, authenticatorISMG)
	}
}

func TestPduConnect_IEncodeFieldLengthError(t *testing.T) {
	c := &PduConnect{
		Header: cmpp.Header{
			CommandID:  cmpp.CommandConnect,
			SequenceID: 1,
		},
		SourceAddr: "1234567",
	}

	data, err := c.IEncode()
	assert.Nil(t, data)
	if !assert.Error(t, err) {
		return
	}

	var lengthErr *packet.FieldLengthError
	if !errors.As(err, &lengthErr) {
		t.Fatalf("errors.As(%T) did not expose FieldLengthError: %v", err, err)
	}
	assert.Equal(t, "SourceAddr", lengthErr.Field)
	assert.Equal(t, 7, lengthErr.Actual)
	assert.Equal(t, 6, lengthErr.Limit)
	assert.ErrorIs(t, err, packet.ErrFieldLengthExceeded)
	assert.Contains(t, err.Error(), `field "SourceAddr"`)
	assert.Contains(t, err.Error(), "actual length 7")
	assert.Contains(t, err.Error(), "limit 6")
}
