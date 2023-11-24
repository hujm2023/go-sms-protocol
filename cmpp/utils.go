package cmpp

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const ConnectTSFormat = "0102150405"

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

// Now ...
func Now() time.Time {
	return time.Now()
}

// GenConnectTimestamp ...
func GenConnectTimestamp(nowFunc func() time.Time) (string, uint32) {
	if nowFunc == nil {
		nowFunc = Now
	}
	t, _ := strconv.Atoi(nowFunc().Format(ConnectTSFormat))
	s := uint32(t)
	return TimeStamp2Str(s), s
}

// TimeStamp2Str converts a timestamp(MMDDHHMMSS) int to a string(10 bytes).
// Right aligned, fill 0 if shorter than 10.
func TimeStamp2Str(t uint32) string {
	return fmt.Sprintf("%010d", t)
}

// GenConnectAuth is used to generate the AuthenticatorSource field in the CMPP CONNECT PDU.
func GenConnectAuth(account string, password string, timestampStr string) []byte {
	md5Bytes := md5.Sum(
		bytes.Join([][]byte{
			[]byte(account),
			make([]byte, 9),
			[]byte(password),
			[]byte(timestampStr),
		},
			nil),
	)
	return md5Bytes[:]
}

// GenConnectRespAuthISMG ...
func GenConnectRespAuthISMG(statusBytes []byte, reqAuth string, password string) []byte {
	m := md5.Sum(bytes.Join([][]byte{
		statusBytes,
		[]byte(reqAuth),
		[]byte(password)},
		nil),
	)
	return m[:]
}
