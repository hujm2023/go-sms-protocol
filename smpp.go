package protocol

import (
	"context"
	"fmt"

	"github.com/hujm2023/go-sms-protocol/datacoding"
	"github.com/hujm2023/go-sms-protocol/datacoding/gsm7encoding"
)

// EncodeSMPPContentAndSplit 发送时SMPP长短信切分
func EncodeSMPPContentAndSplit(ctx context.Context, content string, msgFmt datacoding.SMPPDataCoding, frameKey byte) (
	contents [][]byte, actualMsgFmt datacoding.SMPPDataCoding, err error,
) {
	actualMsgFmt = msgFmt
	if msgFmt == datacoding.SMPP_CODING_GSM7_PACKED {
		// fast path: content中包含非 gsm7 的编码，没必要再去尝试一遍。直接使用 ucs2
		if datacoding.CanEncodeByGSM7(content) {
			// gsm7特殊处理
			contents, actualMsgFmt, err = encodeAndSplitGSM7Packed(content, frameKey)
			if err == nil {
				return contents, actualMsgFmt, nil
			}
			// log.V1.CtxNotice(ctx, "[EncodeSMPPContentAndSplit] content encode with gsm7 error: %v. use ucs2")
		}
		actualMsgFmt = datacoding.SMPP_CODING_UCS2
	}

	var encodedData []byte
	encoder := datacoding.GetSMPPCodec(actualMsgFmt, content)
	encodedData, err = encoder.Encode()
	if err != nil && encoder.Name() != datacoding.DataCodingUcs2 {
		// ucs2兜底
		actualMsgFmt = datacoding.SMPP_CODING_UCS2
		encoder = datacoding.UCS2(content)
		encodedData, err = encoder.Encode()
	}
	if err != nil {
		return nil, 0, fmt.Errorf("encode by %s error: %w", actualMsgFmt, err)
	}

	maxLongLength, perMsgLength := encoder.SplitBy()
	// 短短信
	if len(encodedData) <= maxLongLength {
		return [][]byte{encodedData}, actualMsgFmt, nil
	}

	return splitWithUDHI(encodedData, perMsgLength, frameKey), actualMsgFmt, nil
}

// DecodeSMPPCContent 解码
func DecodeSMPPCContent(ctx context.Context, source string, dataCoding int) (newContent string, err error) {
	var dataBytes []byte
	switch dataCoding {
	case datacoding.SMPP_CODING_GSM7_UNPACKED.ToInt(), datacoding.SMPP_CODING_GSM7_PACKED.ToInt():
		// 先用unpacked去 decode，失败后再次尝试使用packed
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

// encodeAndSplitGSM7Packed 专门处理 gsm7(packed)
// 先 encode，再根据 153 去切分，再 pack 成 134，加上长短信头后返回
func encodeAndSplitGSM7Packed(content string, frameKey byte) ([][]byte, datacoding.SMPPDataCoding, error) {
	dataCoding := datacoding.SMPP_CODING_GSM7_PACKED

	contentBytes, err := gsm7encoding.Encode(content)
	if err != nil {
		return nil, 0, fmt.Errorf("encode error: %w", err)
	}

	// 短短信，直接 pack 后返回
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

		// 边界case：非最后一个 part 的最后一个字节，刚好是拓展字符的标志，若在此处切割，会将这两个字节分开。
		// 为避免这种 case，让前一个 part 少打包一个字节，将0x1b放到下一个字节中
		if idx != msgCount-1 && contentBytes[end-1] == gsm7encoding.EscapeSequence {
			end--
		}

		// 拼接长短信头 + packed 之后 的内容
		contentByte := make([]byte, 0, (end-begin)+datacoding.UDHILength)
		contentByte = append(contentByte, LongMsgHeader6ByteFrameKey)
		contentByte = append(contentByte, LongMsgHeader6ByteFrameTotal)
		contentByte = append(contentByte, LongMsgHeader6ByteFrameNum)
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
