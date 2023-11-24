package cmpp

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseSignature(t *testing.T) {
	for _, item := range []struct {
		content           string
		expectedSignature string
	}{
		{content: "【签名】你的验证码是 123456", expectedSignature: "签名"},
	} {
		assert.Equal(t, item.expectedSignature, ParseSignature(item.content))
	}
}

func TestRemoveSign(t *testing.T) {
	for _, item := range []struct {
		content       string
		resultContent string
		signature     string
	}{
		{content: "", resultContent: "", signature: ""},
		{content: "【a】", resultContent: "【a】", signature: ""},
		{content: "【【【", resultContent: "【【【", signature: ""},
		{content: "】】】", resultContent: "】】】", signature: ""},
		{content: "【【】", resultContent: "【【】", signature: ""},
		{content: "ha【】】", resultContent: "ha【】】", signature: ""},
		{content: "【ha【】】", resultContent: "【ha【】】", signature: ""},

		{content: "你的验证码是 123456", resultContent: "你的验证码是 123456", signature: ""},
		{content: "你的验证码是 【123456】中间的不算签名", resultContent: "你的验证码是 【123456】中间的不算签名", signature: ""},

		{content: "【签名】你的验证码是 123456", resultContent: "你的验证码是 123456", signature: "签名"},                           // normal
		{content: "【签名]你的验证码是 123456", resultContent: "【签名]你的验证码是 123456", signature: ""},                         // 签名格式不对
		{content: "【签名】你的验证码是 123456【另一个签名】", resultContent: "你的验证码是 123456【另一个签名】", signature: "签名"},             // 内容中也出现签名格式，在结尾
		{content: "【签名】你的验证码是【1234】，3 分钟有效", resultContent: "你的验证码是【1234】，3 分钟有效", signature: "签名"},               // 内容中也出现签名格式，在内容中
		{content: "【签名你的验证码是 123456【另一个签名】", resultContent: "【签名你的验证码是 123456【另一个签名】", signature: ""},             // 非标准格式
		{content: "【签名你的验证码是 123456【另一个签名】，3分钟有效", resultContent: "【签名你的验证码是 123456【另一个签名】，3分钟有效", signature: ""}, // 不标准的格式
		{content: "[签名你的验证码是 123456[另一个签名]，3分钟有效", resultContent: "[签名你的验证码是 123456[另一个签名]，3分钟有效", signature: ""}, // 不标准的格式

		{content: "[签名]你的验证码是 123456", resultContent: "你的验证码是 123456", signature: "签名"}, // normal

		{content: "你的验证码是 123456[签名]", resultContent: "你的验证码是 123456", signature: "签名"}, // normal
		{content: "你的验证码是 123456[签名]oha]", resultContent: "你的验证码是 123456[签名]oha]", signature: ""},
	} {
		newContent, signature := RemoveSign(item.content)
		t.Log(newContent, signature)
		assert.Equal(t, item.signature, signature)
		assert.Equal(t, item.resultContent, newContent)
	}
}

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
