package cmpp30

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/hujm2023/go-sms-protocol/cmpp"
)

type ConnectTestSuite struct {
	suite.Suite
	user       string
	version    uint8
	timestamp  string
	sequenceID uint32
	password   string

	nowFunc func() time.Time

	valueBytes []byte
}

func (s *ConnectTestSuite) SetupTest() {
	s.user = "900001"
	s.version = 0x30
	s.sequenceID = 0x17
	s.password = "888888"
	s.nowFunc = func() time.Time {
		return time.Date(2021, 10, 21, 8, 5, 10, 0, time.Local)
	}

	s.valueBytes = []byte{
		0x00, 0x00, 0x00, 0x27, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x17, 0x39, 0x30, 0x30, 0x30,
		0x30, 0x31, 0x90, 0xd0, 0x0c, 0x1d, 0x51, 0x7a, 0xbd, 0x0b, 0x4f, 0x65, 0xf6, 0xbc, 0xf8, 0x53,
		0x5d, 0x16, 0x30, 0x3c, 0xdc, 0x73, 0xbe,
	}
}

func (s *ConnectTestSuite) TestConnect_IDecode() {
	a, t := cmpp.GenConnectTimestamp(s.nowFunc)
	c := &Connect{
		Header: cmpp.Header{
			CommandID:  cmpp.CommandConnect,
			SequenceID: s.sequenceID,
		},
		SourceAddr: s.user,
		// AuthenticatorSource: string(s.authBytes),
		Version:   s.version,
		Timestamp: t,
	}
	c.AuthenticatorSource = string(cmpp.GenConnectAuth(c.SourceAddr, s.password, a))
	data, err := c.IEncode()

	s.Nil(err)
	s.Equal(s.valueBytes, data)
}

func (s *ConnectTestSuite) TestConnect_IEncode() {
	c := new(Connect)
	s.Nil(c.IDecode(s.valueBytes))
	s.Equal(s.user, c.SourceAddr)
	s.Equal(s.version, c.Version)
	a, t := cmpp.GenConnectTimestamp(s.nowFunc)
	s.Equal(t, c.Timestamp)

	s.Equal(string(cmpp.GenConnectAuth(c.SourceAddr, s.password, a)), c.AuthenticatorSource)
}

func TestConnect(t *testing.T) {
	suite.Run(t, new(ConnectTestSuite))
}

type ConnectRespTestSuite struct {
	suite.Suite
	auth       []byte
	secret     string
	valueBytes []byte
}

func (s *ConnectRespTestSuite) SetupTest() {
	s.secret = "888888"
	s.auth = []byte{
		0x90, 0xd0, 0x0c, 0x1d, 0x51, 0x7a, 0xbd, 0x0b,
		0x4f, 0x65, 0xf6, 0xbc, 0xf8, 0x53, 0x5d, 0x16,
	}

	s.valueBytes = []byte{
		0x00, 0x00, 0x00, 0x21, 0x80, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x17, 0x00, 0x00, 0x00, 0x00,
		0x79, 0x42, 0x97, 0x72, 0x74, 0x09, 0x8c, 0xf2, 0x10, 0xab, 0x0c, 0x16, 0xc3, 0x67, 0xbc, 0x8d,
		0x30,
	}
}

func (s *ConnectRespTestSuite) TestConnectResp_IDecode() {
	c := ConnectResp{
		Header: cmpp.Header{
			CommandID:  cmpp.CommandConnectResp,
			SequenceID: 0x17,
		},
		Status:  0,
		Version: 0x30,
	}
	statusBuf := new(bytes.Buffer)
	s.Nil(binary.Write(statusBuf, binary.BigEndian, c.Status))

	c.AuthenticatorISMG = string(cmpp.GenConnectRespAuthISMG(statusBuf.Bytes(), string(s.auth), s.secret))

	data, err := c.IEncode()
	s.Nil(err)
	s.Equal(s.valueBytes, data)
}

func (s *ConnectRespTestSuite) TestConnectResp_IEncode() {
	c := new(ConnectResp)
	s.Nil(c.IDecode(s.valueBytes))

	s.Equal(cmpp.CommandConnectResp, c.Header.CommandID)
	s.Equal(uint32(0x17), c.Header.SequenceID)
	s.Equal(uint32(0), c.Status)
	s.Equal(uint8(0x30), c.Version)

	statusBuf := new(bytes.Buffer)
	s.Nil(binary.Write(statusBuf, binary.BigEndian, c.Status))

	m := cmpp.GenConnectRespAuthISMG(statusBuf.Bytes(), string(s.auth), s.secret)
	s.Equal(string(m), c.AuthenticatorISMG)
}

func TestConnectResp(t *testing.T) {
	suite.Run(t, new(ConnectRespTestSuite))
}
