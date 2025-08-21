package protocol

import (
	"context"
)

func DecodeCMPPContentSimple(ctx context.Context, dataCoding uint8, msgContent []byte) (content []byte, err error) {
	_, _, _, newContent, _ := ParseLongSmsContentBytes(msgContent)
	return DecodeCMPPCContentBytes(ctx, newContent, dataCoding)
}

func DecodeSMPPContentSimple(ctx context.Context, dataCoding uint8, msgContent []byte) (content []byte, err error) {
	_, _, _, newContent, _ := ParseLongSmsContentBytes(msgContent)
	return DecodeSMPPCContentBytes(ctx, newContent, int(dataCoding))
}

func DecodeSGIPContentSimple(ctx context.Context, dataCoding uint8, msgContent []byte) (content []byte, err error) {
	_, _, _, newContent, _ := ParseLongSmsContentBytes(msgContent)
	return DecodeCMPPCContentBytes(ctx, newContent, dataCoding)
}

func DecodeSMGPContentSimplt(ctx context.Context, dataCoding uint8, msgContent []byte) (content []byte, err error) {
	_, _, _, newContent, _ := ParseLongSmsContentBytes(msgContent)
	return DecodeCMPPCContentBytes(ctx, newContent, dataCoding)
}
