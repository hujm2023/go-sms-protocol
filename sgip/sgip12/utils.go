package sgip12

import (
	"time"

	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/sgip"
)

// NewBind ...
func NewBind(account, passwd string, nodeID, seqID uint32) *Bind {
	connectPdu := &Bind{
		Header: sgip.Header{
			TotalLength: 0,
			CommandID:   sgip.SGIP_BIND,
			Sequence:    [3]uint32{nodeID, sgip.Timestamp(time.Now()), seqID},
		},
		Type:     sgip.SP_SMG,
		Name:     account,
		Password: passwd,
	}
	return connectPdu
}

// DecodeSGIP12 解析对应 sgip12 指令
func DecodeSGIP12(data []byte) (sms.PDU, error) {
	header, err := sgip.PeekHeader(data)
	if err != nil {
		return nil, err
	}

	var pdu sms.PDU
	switch header.CommandID {
	case sgip.SGIP_BIND:
		pdu = new(Bind)
	case sgip.SGIP_BIND_REP:
		pdu = new(BindResp)
	case sgip.SGIP_UNBIND:
		pdu = new(Unbind)
	case sgip.SGIP_SUBMIT:
		pdu = new(Submit)
	case sgip.SGIP_SUBMIT_REP:
		pdu = new(SubmitResp)
	case sgip.SGIP_REPORT:
		pdu = new(Report)
	case sgip.SGIP_REPORT_REP:
		pdu = new(ReportResp)
	case sgip.SGIP_DELIVER:
		pdu = new(Deliver)
	case sgip.SGIP_DELIVER_REP:
		pdu = new(DeliverResp)
	}

	if pdu == nil {
		return nil, sms.ErrUnsupportedPacket
	}

	if err = pdu.IDecode(data); err != nil {
		return nil, err
	}
	return pdu, nil
}
