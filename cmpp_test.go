package protocol

import (
	"context"
	"testing"

	"github.com/hujm2023/go-sms-protocol/datacoding"
	"github.com/stretchr/testify/assert"
)

func TestAAA(t *testing.T) {
	s := `	【Keep】您的[退货|换货|寄修单]申请已[完成]，工单号[TH************]，物流报销码[****]，退款链接：https://keep-staticserver.recloud.com.cn/#/keep/keeppage/service-checkreturnfreight/*******a-***a-*a*e-****-*********b**，请按照格式填写与 提交，信息确认无误* - **个工作日内打款`
	res, kk, err := EncodeCMPPContentAndSplit(context.Background(), s, datacoding.CMPP_CODING_UCS2, 123)
	assert.Nil(t, err)
	t.Log(kk.String())
	for _, v := range res {
		t.Log(v)
		t.Log(DecodeCMPPCContent(context.Background(), string(v[6:]), datacoding.CMPP_CODING_UCS2.ToUint8()))
	}
}
