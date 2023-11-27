package smpp50

import (
	protocol "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/smpp"
)

func DecodeSMPP50(data []byte) (protocol.PDU, error) {
	header, err := smpp.PeekHeader(data)
	if err != nil {
		return nil, err
	}

	var pdu protocol.PDU
	switch header.ID {
	case smpp.ENQUIRE_LINK:
		pdu = new(EnquireLink)
	case smpp.ENQUIRE_LINK_RESP:
		pdu = new(EnquireLinkResp)
	case smpp.SUBMIT_SM:
		pdu = new(SubmitSm)
	case smpp.SUBMIT_SM_RESP:
		pdu = new(SubmitSmResp)
	case smpp.DELIVER_SM:
		pdu = new(DeliverSm)
	case smpp.DELIVER_SM_RESP:
		pdu = new(DeliverSmResp)
	case smpp.BIND_TRANSCEIVER, smpp.BIND_RECEIVER, smpp.BIND_TRANSMITTER:
		pdu = new(Bind)
	case smpp.BIND_TRANSCEIVER_RESP, smpp.BIND_RECEIVER_RESP, smpp.BIND_TRANSMITTER_RESP:
		pdu = new(BindResp)
	case smpp.UNBIND:
		pdu = new(Unbind)
	case smpp.GENERIC_NACK:
		pdu = new(GenericNack)
	}

	if pdu == nil {
		return nil, protocol.ErrUnsupportedPacket
	}

	err = pdu.IDecode(data)
	if err != nil {
		return nil, err
	}

	return pdu, nil
}
