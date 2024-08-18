package cmpp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/valyala/bytebufferpool"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

type FuncWithError func() error

var (
	ErrInvalidUtf8Rune       = errors.New("not valid utf8 runes")
	ErrUnsupportedDataCoding = errors.New("unsupported data coding")
)

const (
	ASCII = uint8(0)
	LATIN = uint8(1)
	UCS2  = uint8(8)
	GBK   = uint8(15)

	// UCS2RemoveSign 有些通道，下游供应商会自动填充签名，此时发送时需要将 content 中的签名去掉
	// 此编码同 UCS2，但对于运营商会自动加签名的通道，此编码可以让运营商不再自动加签名
	UCS2RemoveSign = uint8(25)
)

func Utf8ToUcs2(in string) (string, error) {
	if !utf8.ValidString(in) {
		return "", ErrInvalidUtf8Rune
	}

	r := bytes.NewReader([]byte(in))
	t := transform.NewReader(r, unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewEncoder()) // UTF-16 bigEndian, no-bom
	out, err := io.ReadAll(t)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func Utf8ToUcs2Back(in string) string {
	buf := utf16.Encode([]rune(in))
	octets := make([]byte, 0, len(buf)*2)
	for _, n := range buf {
		octets = append(octets, byte(n&0xFF00>>8), byte(n&0x00FF))
	}
	return string(octets)
}

var ucs2BytesBufferPool = bytebufferpool.Pool{}

func Utf8ToUcs2Pooled(in string) (s string) {
	buf := utf16.Encode([]rune(in))
	octets := ucs2BytesBufferPool.Get()
	for _, n := range buf {
		_ = octets.WriteByte(byte(n & 0xFF00 >> 8))
		_ = octets.WriteByte(byte(n & 0x00FF))
	}
	s = octets.String()
	octets.Reset()
	ucs2BytesBufferPool.Put(octets)
	return s
}

// RemoveSign 移除发送内容中的签名。
// - 成功移除时，signature不为空，返回去除签名后的content和signature；
// - 不符合签名规则，返回原内容
func RemoveSign(sourceContent string) (newContent string, signature string) {
	if sourceContent == "" {
		return sourceContent, ""
	}
	content := []rune(sourceContent)

	if len(content) <= 3 {
		return sourceContent, ""
	}

	start, end, lastIndex := 0, 0, len(content)-1

	var signInSuffix bool
	var invalid bool

	switch {
	case content[0] == '【':
		end = findRune(content, '】')
		// 避免出现`【【】` 场景
		if nextStart := findRune(content[1:], '【'); nextStart > start && nextStart < end {
			invalid = true
		}
	case content[0] == '[':
		end = findRune(content, ']')
		if nextStart := findRune(content[1:], '['); nextStart > start && nextStart < end {
			invalid = true
		}
	case content[lastIndex] == '】':
		signInSuffix = true
		start = findRuneReverse(content[:lastIndex-1], '【')
		// `【】】`场景
		if prevEnd := findRuneReverse(content[:lastIndex], '】'); start == -1 || (prevEnd > 0 && prevEnd > start) {
			invalid = true
		}
	case content[lastIndex] == ']':
		signInSuffix = true
		start = findRune(content, '[')
		if prevEnd := findRuneReverse(content[:lastIndex-1], ']'); start == -1 || (prevEnd > 0 && prevEnd > start) {
			invalid = true
		}
	default:
		invalid = true
	}

	if invalid {
		return sourceContent, ""
	}

	// 签名前置(前后都有签名时，优先考虑前置的签名)
	if !signInSuffix && end > 1 && end < lastIndex {
		return string(content[end+1:]), string(content[start+1 : end])
	}

	// 签名后置
	if signInSuffix && start > 0 && start < lastIndex {
		return string(content[:start]), string(content[start+1 : lastIndex])
	}

	// 默认(没签名、不是标准的签名、去除签名后没内容)
	return sourceContent, ""
}

func findRune(content []rune, r rune) int {
	if len(content) == 0 {
		return -1
	}

	for idx, v := range content {
		if v == r {
			return idx
		}
	}

	return -1
}

func findRuneReverse(content []rune, r rune) int {
	if len(content) == 0 {
		return -1
	}

	for idx := len(content) - 1; idx >= 0; idx-- {
		if content[idx] == r {
			return idx
		}
	}

	return -1
}

// ParseSignature 解析内容中的签名
func ParseSignature(content string) string {
	signature := signaturePosition(content, "【", "】")
	if signature != "" {
		return signature
	}
	return signaturePosition(content, "[", "]")
}

func signaturePosition(content string, left, right string) string {
	from := strings.Index(content, left)
	if from < 0 {
		return ""
	}

	to := strings.Index(content, right)
	if to < 0 {
		return ""
	}

	if from >= to || to <= len(left) {
		return ""
	}

	return content[from+len(left) : to]
}

// 1 字节，状态：0正确 1消息结构错 2非法源地址 3认证错 4版本太高 >5其他错误
var connectRespStatus = map[uint8]string{
	0: "正确",
	1: "消息结构错",
	2: "非法源地址",
	3: "认证错",
	4: "版本太高",
	5: "其他错误",
}

func ConnectRespResultString(r uint8) string {
	if s, ok := connectRespStatus[r]; ok {
		return s
	}
	return fmt.Sprintf("未知错误代码 %d", r)
}
