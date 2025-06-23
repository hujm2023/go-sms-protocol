package protocol

import (
	"context"
)

func DecodeCMPPContentSimple(ctx context.Context, dataCoding uint8, msgContent string) (content string, err error) {
	_, _, _, newContent, _ := ParseLongSmsContent(msgContent)
	return DecodeCMPPCContent(ctx, newContent, dataCoding)
}

func DecodeSMPPContentSimple(ctx context.Context, dataCoding uint8, msgContent string) (content string, err error) {
	_, _, _, newContent, _ := ParseLongSmsContent(msgContent)
	return DecodeSMPPCContent(ctx, newContent, int(dataCoding))
}

func DecodeSGIPContentSimple(ctx context.Context, dataCoding uint8, msgContent string) (content string, err error) {
	_, _, _, newContent, _ := ParseLongSmsContent(msgContent)
	return DecodeCMPPCContent(ctx, newContent, dataCoding)
}

func DecodeSGIPContent(ctx context.Context, dataCoding int, msgContent string) (content string, err error) {
	_, _, _, newContent, _ := ParseLongSmsContent(msgContent)
	return DecodeCMPPCContent(ctx, newContent, uint8(dataCoding))
}
