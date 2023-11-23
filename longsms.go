package protocol

import "github.com/hujm2023/go-sms-protocol/datacoding"

const (
	MinLongSmsHeaderLength  = 6   // 长短信额头部长度
	MaxLongSmsContentLength = 140 // 非GSM7编码的短信最大长度

	LongMsgHeader6ByteFrameKey   = byte(0x05) // 长短信6位格式协议头 05 00 03 XX MM NN
	LongMsgHeader6ByteFrameTotal = byte(0x00)
	LongMsgHeader6ByteFrameNum   = byte(0x03)
	Default6FrameKey             = 107

	LongMsgHeader7ByteFrameKey   = byte(0x06) // 长短信7位格式协议头 06 08 04 XX XX MM NN
	LongMsgHeader7ByteFrameTotal = byte(0x08)
	LongMsgHeader7ByteFrameNum   = byte(0x04)
)

// ParseLongSmsContent 解析长短信头
// frameKey: 这批短信唯一标识
// total: 这批长短信总数量
// index: 此条短信在这批长短信中的 index，从 1 开始
// newContent: 此条短信中去除长短信头后的真实内容
// valid: 是否符合长短信格式
func ParseLongSmsContent(content string) (frameKey, total, index int, newContent string, valid bool) {
	newContent = content
	valid = true
	if len(content) < MinLongSmsHeaderLength {
		valid = false
		return
	}
	switch {
	case content[0] == LongMsgHeader6ByteFrameKey && content[1] == LongMsgHeader6ByteFrameTotal && content[2] == LongMsgHeader6ByteFrameNum:
		frameKey = int(content[3] & 0xff)
		total = int(content[4] & 0xff)
		index = int(content[5] & 0xff)
		newContent = content[6:]
	case len(content) > MinLongSmsHeaderLength &&
		content[0] == LongMsgHeader7ByteFrameKey && content[1] == LongMsgHeader7ByteFrameTotal && content[2] == LongMsgHeader7ByteFrameNum:
		frameKey = int((content[3] & 0xff) | (content[4] & 0xff))
		total = int(content[5] & 0xff)
		index = int(content[6] & 0xff)
		newContent = content[7:]
	default:
		// 其他格式不支持，或者不是长短信标准格式
		valid = false
	}
	return
}

// splitWithUDHI 根据 perMsgLength 切分长短信，并添加6位长短信头
func splitWithUDHI(data []byte, perMsgLength int, frameKey byte) [][]byte {
	total := len(data)
	msgCount := ceil(total, perMsgLength) // 向上取整
	contentBytes := make([][]byte, 0, msgCount)
	for idx := 0; idx < msgCount; idx++ {
		contentByte := make([]byte, 0, perMsgLength+datacoding.UDHILength) // 一次性申请足够多空间的内存，避免 append 时发生 copy

		// 长短信头
		contentByte = append(contentByte, LongMsgHeader6ByteFrameKey)
		contentByte = append(contentByte, LongMsgHeader6ByteFrameTotal)
		contentByte = append(contentByte, LongMsgHeader6ByteFrameNum)
		contentByte = append(contentByte, frameKey)       // frameKey
		contentByte = append(contentByte, byte(msgCount)) // total
		contentByte = append(contentByte, byte(idx+1))    // num

		// 按照 perMsgLength 切割
		begin := idx * perMsgLength
		end := (idx + 1) * perMsgLength
		if end > total {
			end = total
		}
		if begin == end {
			continue
		}
		contentByte = append(contentByte, data[begin:end]...)

		contentBytes = append(contentBytes, contentByte)
	}

	return contentBytes
}

// ceil 向上取整
func ceil(total, split int) int {
	return (total + split - 1) / split
}
