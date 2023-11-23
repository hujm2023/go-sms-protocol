package protocol

import (
	"bytes"
	"context"
	"testing"

	"github.com/hujm2023/go-sms-protocol/datacoding"
	"github.com/stretchr/testify/assert"
)

func TestSplitLongContent(t *testing.T) {
	content := "【德邦快递】尊敬客户您好，您反馈的问题将由我（工号：7142）负责专职跟进，问题目前正在处理中 在此期间，您不需要进行任何操作，有最新处理进展我们将及时联系您，请保持电话畅通。期间如有问题咨询可回拨95353热线1-4号键转工号7142，会由我来回复您，祝您生活愉快！"
	data, coding, err := EncodeCMPPContentAndSplit(context.Background(), content, datacoding.CMPP_CODING_UCS2, 123)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, datacoding.CMPP_CODING_UCS2, coding)
	assert.Equal(t, 2, len(data))
	for idx, cc := range data {
		assert.True(t, bytes.Equal([]byte{0x05, 0x00, 0x03, 123, 2, byte(idx + 1)}, []byte(cc[:6])))
	}
}

func TestCeil(t *testing.T) {
	for _, i := range []struct {
		total  int
		split  int
		result int
	}{
		{total: 142, split: 134, result: 2},
		{total: 268, split: 134, result: 2},
		{total: 161, split: 153, result: 2},
		{total: 120, split: 134, result: 1},
		{total: 1, split: 134, result: 1},
	} {
		if v := ceil(i.total, i.split); v != i.result {
			t.Fatalf("[ceil] failed. total: %d, splitWithUDHI: %d, want: %d, actial: %d", i.total, i.split, i.result, v)
		}
	}
}
