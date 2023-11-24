package cmpp30

import (
	protocol "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/cmpp"
)

// DecodeCMPP30 ...
func DecodeCMPP30(data []byte) (protocol.PDU, error) {
	header, err := cmpp.PeekHeader(data)
	if err != nil {
		return nil, err
	}

	var pdu protocol.PDU
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
		return nil, protocol.ErrUnsupportedPacket
	}

	if err = pdu.IDecode(data); err != nil {
		return nil, err
	}
	return pdu, nil
}
