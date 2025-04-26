package cmpp30

import (
	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/cmpp"
)

// DecodeCMPP30 decodes the given byte slice into a corresponding CMPP 3.0 PDU.
func DecodeCMPP30(data []byte) (sms.PDU, error) {
	header, err := cmpp.PeekHeader(data)
	if err != nil {
		return nil, err
	}

	var pdu sms.PDU
	switch header.CommandID {
	case cmpp.CommandActiveTest:
		pdu = new(ActiveTest)
	case cmpp.CommandActiveTestResp:
		pdu = new(ActiveTestResp)
	case cmpp.CommandSubmit:
		pdu = new(Submit)
	case cmpp.CommandSubmitResp:
		pdu = new(SubmitResp)
	case cmpp.CommandDeliver:
		pdu = new(Deliver)
	case cmpp.CommandDeliverResp:
		pdu = new(DeliverResp)
	case cmpp.CommandTerminate:
		pdu = new(Terminate)
	case cmpp.CommandTerminateResp:
		pdu = new(TerminateResp)
	case cmpp.CommandConnect:
		pdu = new(Connect)
	case cmpp.CommandConnectResp:
		pdu = new(ConnectResp)
	case cmpp.CommandQuery:
		pdu = new(Query)
	case cmpp.CommandQueryResp:
		pdu = new(QueryResp)
	case cmpp.CommandCancel:
		pdu = new(Cancel)
	case cmpp.CommandCancelResp:
		pdu = new(CancelResp)
	}

	if pdu == nil {
		return nil, sms.ErrUnsupportedPacket
	}

	if err = pdu.IDecode(data); err != nil {
		return nil, err
	}
	return pdu, nil
}
