package cmpp

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
)

var _testUTF8Content = "123;hello world;你好;안녕하세요;こんにちは;ON LI DAY FaOHE MASHI;hallo! Wie geht es dir;bonjour;moiẽn;hallo;Olá"

func TestCompareUtf8ToUcs2(t *testing.T) {
	for _, s := range []string{
		"123",
		"hello world",
		"你好",
		_testUTF8Content,
		"رمز التحقق الخاص بك هو 1234",
	} {
		v1, err := Utf8ToUcs2(s)
		assert.Nil(t, err)
		v2 := Utf8ToUcs2Back(s)
		v3 := Utf8ToUcs2Pooled(s)
		assert.Equal(t, v1, v2)
		assert.Equal(t, v1, v3)
		assert.Equal(t, v2, v3)
		t.Log([]byte(v1))
	}
}

func BenchmarkUtf8ToUcs2(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Utf8ToUcs2(_testUTF8Content)
	}
}

func BenchmarkUtf8ToUcs2Back(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Utf8ToUcs2Back(_testUTF8Content)
	}
}

func BenchmarkUtf8ToUcs2Pooled(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Utf8ToUcs2Pooled(_testUTF8Content)
	}
}

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

func TestA(t *testing.T) {
	b := []byte{0, 0, 1, 49}
	a := binary.BigEndian.Uint32(b)
	t.Log(b, a)
}
