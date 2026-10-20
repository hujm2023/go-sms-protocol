package cmpp20

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"testing"

	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/packet"
	"github.com/stretchr/testify/assert"
)

func TestNewConnect(t *testing.T) {
	c := NewConnect("900001", "888888", 0x17)
	data, err := c.IEncode()
	assert.Nil(t, err)
	assert.Equal(t, "900001", c.SourceAddr)
	assert.Equal(t, cmpp.CommandConnect, c.CommandID)
	assert.Equal(t, uint32(0x17), c.SequenceID)
	assert.Equal(t, uint8(cmpp.Version20), c.Version)
	assert.Len(t, data, MaxConnectLength)
	assert.Equal(t, uint32(len(data)), binary.BigEndian.Uint32(data[:cmpp.HeaderLength]))
}

func TestNewConnectAuthenticator(t *testing.T) {
	for _, test := range []struct {
		name     string
		account  string
		password string
	}{
		{name: "short account and short password", account: "90001", password: "abc"},
		{name: "short account and six byte password", account: "90001", password: "abcdef"},
		{name: "short account and long password", account: "90001", password: "abcdefgh"},
		{name: "six byte account and short password", account: "900001", password: "abc"},
		{name: "six byte account and six byte password", account: "900001", password: "abcdef"},
		{name: "six byte account and long password", account: "900001", password: "abcdefgh"},
	} {
		t.Run(test.name, func(t *testing.T) {
			connect := NewConnect(test.account, test.password, 0x17)
			data, err := connect.IEncode()
			if err != nil {
				t.Fatalf("IEncode() error = %v", err)
			}
			if len(data) != MaxConnectLength {
				t.Fatalf("encoded length = %d, want %d", len(data), MaxConnectLength)
			}

			// Use the encoded SourceAddr as an independent authentication input.
			input := append([]byte{}, data[cmpp.HeaderLength:cmpp.HeaderLength+6]...)
			input = append(input, make([]byte, 9)...)
			input = append(input, []byte(test.password)...)
			input = append(input, []byte(cmpp.TimeStamp2Str(connect.Timestamp))...)
			expected := md5.Sum(input)
			auth := data[cmpp.HeaderLength+6 : cmpp.HeaderLength+6+md5.Size]
			if !bytes.Equal(auth, expected[:]) {
				t.Fatalf("encoded AuthenticatorSource = %x, want %x", auth, expected)
			}
			if connect.SourceAddr != test.account {
				t.Fatalf("SourceAddr = %q, want %q", connect.SourceAddr, test.account)
			}
		})
	}
}

func TestNewConnectRejectsLongAccount(t *testing.T) {
	connect := NewConnect("1234567", "abcdef", 0x17)
	data, err := connect.IEncode()
	if data != nil {
		t.Fatalf("IEncode() data = %x, want nil", data)
	}
	var lengthErr *packet.FieldLengthError
	if !errors.As(err, &lengthErr) {
		t.Fatalf("IEncode() error = %v, want FieldLengthError", err)
	}
	if lengthErr.Field != "SourceAddr" {
		t.Fatalf("field length error field = %q, want SourceAddr", lengthErr.Field)
	}
	if lengthErr.Actual != 7 || lengthErr.Limit != 6 {
		t.Fatalf("unexpected field length error: %+v", lengthErr)
	}
}

func TestDecodeCMPP20_UnsupportedCommand(t *testing.T) {
	raw := []byte{0, 0, 0, 12, 0xde, 0xad, 0xbe, 0xef, 0, 0, 0, 1}
	if _, err := DecodeCMPP20(raw); !errors.Is(err, sms.ErrUnsupportedPacket) {
		t.Fatalf("DecodeCMPP20() error = %v, want %v", err, sms.ErrUnsupportedPacket)
	}
}
