package protocol

import (
	"context"
	"testing"

	"github.com/hujm2023/go-sms-protocol/datacoding"
	"github.com/stretchr/testify/suite"
)

// ContentTestSuite is a test suite for content decoding functions.
type ContentTestSuite struct {
	suite.Suite
}

// TestContentTestSuite runs the ContentTestSuite.
func TestContentTestSuite(t *testing.T) {
	suite.Run(t, new(ContentTestSuite))
}

// TestDecodeCMPPContentSimple tests the DecodeCMPPContentSimple function.
func (s *ContentTestSuite) TestDecodeCMPPContentSimple() {
	tests := []struct {
		name        string
		dataCoding  uint8
		msgContent  string
		wantContent string
		wantErr     bool
	}{
		{
			name:        "Short message ASCII",
			dataCoding:  datacoding.CMPP_CODING_ASCII.ToUint8(),
			msgContent:  "Hello",
			wantContent: "Hello",
			wantErr:     false,
		},
		{
			name:       "Long message header GBK",
			dataCoding: datacoding.CMPP_CODING_GBK.ToUint8(),
			// Long message header (6 bytes) + GBK for "你好"
			msgContent:  string([]byte{0x05, 0x00, 0x03, 0x01, 0x02, 0x01}) + string([]byte{0xc4, 0xe3, 0xba, 0xc3}),
			wantContent: "你好",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			content, err := DecodeCMPPContentSimple(context.Background(), tt.dataCoding, tt.msgContent)
			if tt.wantErr {
				s.Error(err)
			} else {
				s.NoError(err)
				s.Equal(tt.wantContent, content, tt.name)
			}
		})
	}
}

// TestDecodeSMPPContentSimple tests the DecodeSMPPContentSimple function.
func (s *ContentTestSuite) TestDecodeSMPPContentSimple() {
	// Note: DecodeSMPPContentSimple does not handle long message splitting internally in the provided snippet.
	// It passes the full msgContent to DecodeSMPPCContent.
	// We will test it as is.
	content, err := DecodeSMPPContentSimple(context.Background(), uint8(datacoding.SMPP_CODING_ASCII), "test")
	s.NoError(err)
	s.Equal("test", content)
}

// TestDecodeSGIPContentSimple tests the DecodeSGIPContentSimple function.
func (s *ContentTestSuite) TestDecodeSGIPContentSimple() {
	content, err := DecodeSGIPContentSimple(context.Background(), uint8(datacoding.CMPP_CODING_GBK), string([]byte{0x05, 0x00, 0x03, 0x01, 0x02, 0x01})+string([]byte{0xc4, 0xe3, 0xba, 0xc3}))
	s.NoError(err)
	s.Equal("你好", content)
}

// TestDecodeSGIPContent tests the DecodeSGIPContent function.
func (s *ContentTestSuite) TestDecodeSGIPContent() {
	content, err := DecodeSMGPContentSimplt(context.Background(), int(datacoding.CMPP_CODING_GBK), string([]byte{0x05, 0x00, 0x03, 0x01, 0x02, 0x01})+string([]byte{0xc4, 0xe3, 0xba, 0xc3}))
	s.NoError(err)
	s.Equal("你好", content)
}
