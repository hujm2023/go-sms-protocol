package cmpp

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeStamp2Str(t *testing.T) {
	for _, item := range []struct {
		s        uint32
		expected string
	}{
		{s: 1021080510, expected: "1021080510"},
		{s: 121080510, expected: "0121080510"},
	} {
		assert.Equal(t, item.expected, TimeStamp2Str(item.s))
	}
}

func TestGenConnectTimestamp(t *testing.T) {
	for _, item := range []struct {
		now          time.Time
		expectString string
		expectUint32 uint32
	}{
		{
			now:          time.Date(2021, 10, 24, 9, 30, 58, 0, time.Local),
			expectString: "1024093058",
			expectUint32: 1024093058,
		},
		{
			now:          time.Date(2021, 9, 24, 9, 30, 58, 0, time.Local),
			expectString: "0924093058",
			expectUint32: 924093058,
		},
	} {
		a, b := GenConnectTimestamp(func() time.Time { return item.now })
		assert.Equal(t, item.expectString, a)
		assert.Equal(t, item.expectUint32, b)
	}
}

func TestGenConnectAuth(t *testing.T) {
	for _, item := range []struct {
		user      string
		password  string
		timestamp string

		expectAuthString []byte
	}{
		{
			user:      "900001",
			password:  "888888",
			timestamp: "1021080510",
			expectAuthString: []byte{
				0x90, 0xd0, 0x0c, 0x1d, 0x51, 0x7a, 0xbd, 0x0b,
				0x4f, 0x65, 0xf6, 0xbc, 0xf8, 0x53, 0x5d, 0x16,
			},
		},
	} {
		auth := GenConnectAuth(item.user, item.password, item.timestamp)
		assert.Equal(t, item.expectAuthString, auth)
	}
}

func TestGenConnectAuthSourceAddrPadding(t *testing.T) {
	for _, test := range []struct {
		name      string
		account   string
		password  string
		timestamp string
		expected  string
	}{
		{
			name:      "five byte account",
			account:   "90001",
			password:  "abcdef",
			timestamp: "1021080510",
			expected:  "87de1145d8aea707539c9d5259fd8932",
		},
		{
			name:      "one byte account",
			account:   "9",
			password:  "abcdef",
			timestamp: "1021080510",
			expected:  "329a7370dc34c33cc1ccb1021c42ba56",
		},
		{
			name:      "short password",
			account:   "90001",
			password:  "abc",
			timestamp: "1021080510",
			expected:  "448495caada4bcce985e50c4e15b84ea",
		},
		{
			name:      "long password",
			account:   "90001",
			password:  "abcdefgh",
			timestamp: "1021080510",
			expected:  "d068b6aabe600081ab8914dbd046152b",
		},
		{
			name:      "six byte account with short password",
			account:   "900001",
			password:  "abc",
			timestamp: "1021080510",
			expected:  "0698554c8c3c601958067183d8fc42f7",
		},
		{
			name:      "timestamp with leading zero",
			account:   "90001",
			password:  "abcdef",
			timestamp: "0121080510",
			expected:  "6564e65ee831d7f2842438d4ca59a670",
		},
		{
			name:      "already padded account",
			account:   "90001\x00",
			password:  "abcdef",
			timestamp: "1021080510",
			expected:  "87de1145d8aea707539c9d5259fd8932",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			auth := GenConnectAuth(test.account, test.password, test.timestamp)
			if got := hex.EncodeToString(auth); got != test.expected {
				t.Fatalf("GenConnectAuth() = %s, want %s", got, test.expected)
			}
		})
	}
}

func TestGenConnectRespAuthISMG(t *testing.T) {
	status2Bytes := func(t *testing.T, status any) []byte {
		t.Helper()
		b := bytes.NewBuffer(nil)
		switch v := status.(type) {
		case uint8:
			assert.Nil(t, binary.Write(b, binary.BigEndian, v))
		case uint32:
			assert.Nil(t, binary.Write(b, binary.BigEndian, v))
		default:
			t.Fatal("invalid status type")
		}
		return b.Bytes()
	}
	for _, item := range []struct {
		status    []byte
		authBytes []byte
		password  string

		expectAuthString []byte
	}{
		{
			status: status2Bytes(t, uint32(0)), // cmpp3.0
			authBytes: []byte{
				0x90, 0xd0, 0x0c, 0x1d, 0x51, 0x7a, 0xbd, 0x0b,
				0x4f, 0x65, 0xf6, 0xbc, 0xf8, 0x53, 0x5d, 0x16,
			},
			password: "888888",
			expectAuthString: []byte{
				0x79, 0x42, 0x97, 0x72, 0x74, 0x09, 0x8c, 0xf2, 0x10, 0xab, 0x0c, 0x16, 0xc3, 0x67, 0xbc, 0x8d,
			},
		},
		{
			status: status2Bytes(t, uint8(0)),
			authBytes: []byte{
				0x90, 0xd0, 0x0c, 0x1d, 0x51, 0x7a, 0xbd, 0x0b,
				0x4f, 0x65, 0xf6, 0xbc, 0xf8, 0x53, 0x5d, 0x16,
			},
			password: "888888",
			expectAuthString: []byte{
				0x6c, 0x0b, 0x84, 0x6e, 0x25, 0xba, 0xb6, 0xda,
				0xa4, 0xed, 0x1c, 0x46, 0x6e, 0x0f, 0x4b, 0xd8,
			},
		},
	} {
		auth := GenConnectRespAuthISMG(item.status, string(item.authBytes), item.password)
		assert.Equal(t, item.expectAuthString, auth)
	}
}
