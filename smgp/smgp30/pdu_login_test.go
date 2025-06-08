package smgp30

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/hujm2023/go-sms-protocol/smgp"
)

type LoginTestSuite struct {
	suite.Suite

	valueBytes []byte
}

func (s *LoginTestSuite) SetupTest() {
	s.valueBytes = []byte{
		0x0, 0x0, 0x0, 0x2a, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0x74, 0x65, 0x73, 0x74, 0x69, 0x64, 0x0, 0x0, 0xca, 0xe, 0xff, 0x3, 0x6f, 0x70, 0xaa, 0x11, 0x16, 0xb8, 0xe5, 0xdf, 0x1d, 0x55, 0xcf, 0x48, 0x0, 0x3c, 0xdc, 0x73, 0xbe, 0x30,
	}
}

func (s *LoginTestSuite) TestLogin_IDecode() {
	c := Login{
		Header:    smgp.Header{TotalLength: 12, CommandID: smgp.CommandLogin, SequenceID: 1},
		ClientID:  "testid",
		LoginMode: smgp.SEND_MODE,
		Version:   0x30,
		Timestamp: genTimestampForTest(),
	}
	au, err := genAuthenticatorClient("testid", "testauth", genTimestampForTest())
	s.Nil(err)
	c.AuthenticatorClient = string(au)
	data, err := c.IEncode()

	s.Nil(err)
	s.Equal(s.valueBytes, data)
}

func (s *LoginTestSuite) TestLogin_IEncode() {
	c := new(Login)
	s.Nil(c.IDecode(s.valueBytes))
	s.Equal(uint8(0x30), c.Version)
	s.Equal("testid", c.ClientID)
	s.Equal(genTimestampForTest(), c.Timestamp)
	au, err := genAuthenticatorClient("testid", "testauth", genTimestampForTest())
	s.Nil(err)
	s.Equal(string(au), c.AuthenticatorClient)
}

func TestLogin(t *testing.T) {
	suite.Run(t, new(LoginTestSuite))
}

type LoginRespTestSuite struct {
	suite.Suite
	auth       []byte
	secret     string
	valueBytes []byte
}

func (s *LoginRespTestSuite) SetupTest() {
	s.valueBytes = []byte{
		0x0, 0x0, 0x0, 0x21, 0x80, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x1, 0x0, 0x0, 0x0, 0x0, 0x41, 0x75, 0x74, 0x68, 0x53, 0x65, 0x72, 0x76, 0x65, 0x72, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x30,
	}
}

func (s *LoginRespTestSuite) TestLoginResp_IEncode() {
	c := LoginResp{
		Header:              smgp.Header{TotalLength: 12, CommandID: smgp.CommandLoginResp, SequenceID: 1},
		Status:              0,
		AuthenticatorServer: "AuthServer",
		ServerVersion:       0x30,
	}

	data, err := c.IEncode()
	s.Nil(err)
	s.Equal(s.valueBytes, data)
}

func (s *LoginRespTestSuite) TestLoginResp_IDecode() {
	c := new(LoginResp)
	s.Nil(c.IDecode(s.valueBytes))

	s.Equal(smgp.CommandLoginResp, c.Header.CommandID)
	s.Equal(uint32(0x1), c.Header.SequenceID)
	s.Equal(LoginRespStatusSuccess, c.Status)
	s.Equal(LoginRespStatusSuccess.String(), c.Status.String())
	s.Equal(uint8(0x30), c.ServerVersion)
	s.Equal("AuthServer", c.AuthenticatorServer)
	s.T().Log(c.String())
}

func TestLoginResp(t *testing.T) {
	suite.Run(t, new(LoginRespTestSuite))
}

func genTimestampForTest() uint32 {
	t := time.Date(2021, 10, 21, 8, 5, 10, 0, time.Local)
	return uint32(int(t.Month())*100000000 + t.Day()*1000000 +
		t.Hour()*10000 + t.Minute()*100 + t.Second())
}
