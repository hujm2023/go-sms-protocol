package protocol

import (
	"context"
	"fmt"

	"github.com/hujm2023/go-sms-protocol/datacoding"
)

// EncodeCMPPContentAndSplit 发送时CMPP长短信切分
func EncodeCMPPContentAndSplit(ctx context.Context, content string, msgFmt datacoding.CMPPDataCoding, frameKey byte) (
	contents [][]byte, actualMsgFmt datacoding.CMPPDataCoding, err error,
) {
	actualMsgFmt = msgFmt
	var encodedData []byte
	encoder := datacoding.GetCMPPCodec(msgFmt, content)
	encodedData, err = encoder.Encode()
	if err != nil && encoder.Name() != datacoding.DataCodingUcs2 {
		// ucs2兜底
		actualMsgFmt = datacoding.CMPP_CODING_UCS2
		encoder = datacoding.UCS2(content)
		encodedData, err = encoder.Encode()
	}
	if err != nil {
		return nil, 0, fmt.Errorf("encode error: %w", err)
	}

	maxLongLength, perMsgLength := encoder.SplitBy()
	// 短短信
	if len(encodedData) <= maxLongLength {
		return [][]byte{encodedData}, actualMsgFmt, nil
	}

	return splitWithUDHI(encodedData, perMsgLength, frameKey), actualMsgFmt, nil
}

// DecodeCMPPCContent 解码
func DecodeCMPPCContent(_ context.Context, source string, dataCoding uint8) (newContent string, err error) {
	var dataBytes []byte
	switch dataCoding {
	case datacoding.CMPP_CODING_ASCII.ToUint8():
		dataBytes, err = datacoding.Ascii(source).Decode()
	case datacoding.CMPP_CODING_GBK.ToUint8():
		dataBytes, err = datacoding.GB18030(source).Decode()
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
