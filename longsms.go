package protocol

import (
	"context"
	"fmt"

	"github.com/hujm2023/go-sms-protocol/datacoding"
	"github.com/hujm2023/go-sms-protocol/datacoding/gsm7encoding"
)

const (
	minLongSmsHeaderLength = 6 // Additional header length for long SMS

	// Protocol header format of 6 for long SMS: 05 00 03 XX MM NN
	longMsgHeader6ByteFrameKey   = byte(0x05)
	longMsgHeader6ByteFrameTotal = byte(0x00)
	longMsgHeader6ByteFrameNum   = byte(0x03)
	default6FrameKey             = 107

	// Protocol header format of 7 for long SMS: 06 08 04 XX XX MM NN
	longMsgHeader7ByteFrameKey   = byte(0x06)
	longMsgHeader7ByteFrameTotal = byte(0x08)
	longMsgHeader7ByteFrameNum   = byte(0x04)
)

// ParseLongSmsContent parses the header of a concatenated SMS.
// frameKey: Unique identifier for this batch of messages
// total: Total count of messages in this concatenated SMS batch
// index: Index of this message within the concatenated SMS batch, starting from 1
// newContent: The actual content of this message after removing the concatenated SMS header
// valid: Indicates whether it conforms to the concatenated SMS format.
func ParseLongSmsContent(content string) (frameKey, total, index int, newContent string, valid bool) {
	newContent = content
	valid = true
	if len(content) < minLongSmsHeaderLength {
		valid = false
		return
	}
	switch {
	case content[0] == longMsgHeader6ByteFrameKey && content[1] == longMsgHeader6ByteFrameTotal && content[2] == longMsgHeader6ByteFrameNum:
		frameKey = int(content[3] & 0xff)
		total = int(content[4] & 0xff)
		index = int(content[5] & 0xff)
		newContent = content[6:]
	case len(content) > minLongSmsHeaderLength &&
		content[0] == longMsgHeader7ByteFrameKey && content[1] == longMsgHeader7ByteFrameTotal && content[2] == longMsgHeader7ByteFrameNum:
		frameKey = int((content[3] & 0xff) | (content[4] & 0xff))
		total = int(content[5] & 0xff)
		index = int(content[6] & 0xff)
		newContent = content[7:]
	default:
		// Not a supported UDHI
		valid = false
	}
	return
}

// EncodeCMPPContentAndSplit encodes CMPP content and, if necessary, performs segmentation for long messages.
// If the provided encoding cannot be applied to the content, UCS2 encoding will be used as a fallback.
// It ultimately returns the encoded binary, the actual encoding used, and potential errors that may occur.
func EncodeCMPPContentAndSplit(ctx context.Context, content string, msgFmt datacoding.CMPPDataCoding, frameKey byte) (
	contents [][]byte, actualMsgFmt datacoding.CMPPDataCoding, err error,
) {
	actualMsgFmt = msgFmt
	var encodedData []byte
	encoder := datacoding.GetCMPPCodec(msgFmt, content)
	encodedData, err = encoder.Encode()
	if err != nil && encoder.Name() != datacoding.DataCodingUcs2 {
		// use ucs2 as fallback
		actualMsgFmt = datacoding.CMPP_CODING_UCS2
		encoder = datacoding.UCS2(content)
		encodedData, err = encoder.Encode()
	}
	if err != nil {
		return nil, 0, fmt.Errorf("encode error: %w", err)
	}

	maxLongLength, perMsgLength := encoder.SplitBy()
	// short message
	if len(encodedData) <= maxLongLength {
		return [][]byte{encodedData}, actualMsgFmt, nil
	}

	return splitWithUDHI(encodedData, perMsgLength, frameKey), actualMsgFmt, nil
}

// DecodeCMPPCContent decodes CMPP content using the provided dataCoding.
func DecodeCMPPCContent(_ context.Context, source string, dataCoding uint8) (newContent string, err error) {
	var dataBytes []byte
	switch dataCoding {
	case datacoding.CMPP_CODING_ASCII.ToUint8():
		dataBytes, err = datacoding.Ascii(source).Decode()
	case datacoding.CMPP_CODING_GBK.ToUint8():
		dataBytes, err = datacoding.GB18030(source).Decode()
	case datacoding.CMPP_CODING_UCS2_NO_SIGN.ToUint8():
		fallthrough
	case datacoding.CMPP_CODING_UCS2.ToUint8():
		dataBytes, err = datacoding.UCS2(source).Decode()
	default:
		err = datacoding.ErrUnsupportedDataCoding
	}
	if err != nil {
		return source, err
	}
	return string(dataBytes), nil
}

// EncodeSMPPContentAndSplit encodes SMPP content and, if necessary, performs segmentation for long messages.
// If the provided encoding cannot be applied to the content, UCS2 encoding will be used as a fallback.
// It ultimately returns the encoded binary, the actual encoding used, and potential errors that may occur.
func EncodeSMPPContentAndSplit(ctx context.Context, content string, msgFmt datacoding.SMPPDataCoding, frameKey byte) (
	contents [][]byte, actualMsgFmt datacoding.SMPPDataCoding, err error,
) {
	actualMsgFmt = msgFmt
	if msgFmt == datacoding.SMPP_CODING_GSM7_PACKED {
		// Fast path: If the content includes non-GSM7 encoding,
		// there's no need to attempt again. Use UCS2 directly.
		if datacoding.CanEncodeByGSM7(content) {
			// gsm7
			contents, actualMsgFmt, err = encodeAndSplitGSM7Packed(content, frameKey)
			if err == nil {
				return contents, actualMsgFmt, nil
			}
		}
		actualMsgFmt = datacoding.SMPP_CODING_UCS2
	}

	var encodedData []byte
	encoder := datacoding.GetSMPPCodec(actualMsgFmt, content)
	encodedData, err = encoder.Encode()
	if err != nil && encoder.Name() != datacoding.DataCodingUcs2 {
		// use ucs2 as default
		actualMsgFmt = datacoding.SMPP_CODING_UCS2
		encoder = datacoding.UCS2(content)
		encodedData, err = encoder.Encode()
	}
	if err != nil {
		return nil, 0, fmt.Errorf("encode by %s error: %w", actualMsgFmt, err)
	}

	maxLongLength, perMsgLength := encoder.SplitBy()
	// short message
	if len(encodedData) <= maxLongLength {
		return [][]byte{encodedData}, actualMsgFmt, nil
	}

	return splitWithUDHI(encodedData, perMsgLength, frameKey), actualMsgFmt, nil
}

// DecodeSMPPCContent decodes SMPP content using the provided dataCoding.
func DecodeSMPPCContent(ctx context.Context, source string, dataCoding int) (newContent string, err error) {
	var dataBytes []byte
	switch dataCoding {
	case datacoding.SMPP_CODING_GSM7_UNPACKED.ToInt(), datacoding.SMPP_CODING_GSM7_PACKED.ToInt():
		// Attempt decoding using 'unpacked' first; if unsuccessful, attempt using 'packed' again.
		dataBytes, err = datacoding.GSM7Unpacked(source).Decode()
		if err != nil {
			// log.V1.CtxInfo(ctx, "GSM7 unpacked decode error: %v, use packed", err)
			dataBytes, err = gsm7encoding.Decode(gsm7encoding.Unpack([]byte(source)))
		}
	case datacoding.SMPP_CODING_ASCII.ToInt():
		dataBytes, err = datacoding.Ascii(source).Decode()
	case datacoding.SMPP_CODING_Latin1.ToInt():
		dataBytes, err = datacoding.Latin1(source).Decode()
	case datacoding.SMPP_CODING_UCS2.ToInt():
		dataBytes, err = datacoding.UCS2(source).Decode()
	default:
		err = datacoding.ErrUnsupportedDataCoding
	}
	if err != nil {
		return source, err
	}
	return string(dataBytes), nil
}

// encodeAndSplitGSM7Packed specifically handles GSM7 (packed).
// It encodes first, then segments based on 153 characters, repacks into 134 characters,
// adds a concatenated SMS header, and returns the result.
func encodeAndSplitGSM7Packed(content string, frameKey byte) ([][]byte, datacoding.SMPPDataCoding, error) {
	dataCoding := datacoding.SMPP_CODING_GSM7_PACKED

	contentBytes, err := gsm7encoding.Encode(content)
	if err != nil {
		return nil, 0, fmt.Errorf("encode error: %w", err)
	}

	// short message
	if len(contentBytes) <= datacoding.MaxGSM7Length {
		return [][]byte{gsm7encoding.Pack(contentBytes)}, dataCoding, nil
	}

	perMsgLength := datacoding.SplitBy153
	msgCount := ceil(len(contentBytes), perMsgLength)
	res := make([][]byte, 0, msgCount)

	begin, end := 0, perMsgLength
	for idx := 0; idx < msgCount; idx++ {
		if end > len(contentBytes) {
			end = len(contentBytes)
		}
		if begin >= end {
			continue
		}

		// Boundary case: When the last byte of a non-final part happens to be the indicator for an extended character,
		// cutting at this point would split these two bytes.
		// To avoid this scenario, the preceding part should pack one byte less, ensuring that 0x1b is placed within the next byte.
		if idx != msgCount-1 && contentBytes[end-1] == gsm7encoding.EscapeSequence {
			end--
		}

		// append UDHI
		contentByte := make([]byte, 0, (end-begin)+datacoding.UDHILength)
		contentByte = append(contentByte, longMsgHeader6ByteFrameKey)
		contentByte = append(contentByte, longMsgHeader6ByteFrameTotal)
		contentByte = append(contentByte, longMsgHeader6ByteFrameNum)
		contentByte = append(contentByte, frameKey)       // frameKey
		contentByte = append(contentByte, byte(msgCount)) // total
		contentByte = append(contentByte, byte(idx+1))    // num

		// pack
		packed := gsm7encoding.Pack(contentBytes[begin:end])
		contentByte = append(contentByte, packed...)

		res = append(res, contentByte)

		begin = end
		end += perMsgLength
	}

	return res, dataCoding, nil
}

// splitWithUDHI splits the long message according to perMsgLength and adds a 6-byte header for concatenated SMS.
func splitWithUDHI(data []byte, perMsgLength int, frameKey byte) [][]byte {
	total := len(data)
	msgCount := ceil(total, perMsgLength)
	contentBytes := make([][]byte, 0, msgCount)
	for idx := 0; idx < msgCount; idx++ {
		contentByte := make([]byte, 0, perMsgLength+datacoding.UDHILength)

		// append UDHI
		contentByte = append(contentByte, longMsgHeader6ByteFrameKey)
		contentByte = append(contentByte, longMsgHeader6ByteFrameTotal)
		contentByte = append(contentByte, longMsgHeader6ByteFrameNum)
		contentByte = append(contentByte, frameKey)       // frameKey
		contentByte = append(contentByte, byte(msgCount)) // total
		contentByte = append(contentByte, byte(idx+1))    // num

		// split by perMsgLength
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

// ceil: rounding up to the nearest integer.
func ceil(total, split int) int {
	return (total + split - 1) / split
}
