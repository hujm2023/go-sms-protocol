package cmpp

import (
	"testing"

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
