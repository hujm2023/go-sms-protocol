package smpp34

import (
	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/smpp"
)

// NewEnquireLinkReqBytes ...
func NewEnquireLinkReqBytes(seqID uint32) []byte {
	b := &EnquireLink{
		Header: smpp.Header{
			ID:       smpp.ENQUIRE_LINK,
			Sequence: seqID,
		},
	}
	data, _ := b.IEncode()
	return data
}

// NewEnquireLinkRespBytes ...
func NewEnquireLinkRespBytes(seqID uint32) []byte {
	b := &EnquireLinkResp{
		Header: smpp.Header{
			ID:       smpp.ENQUIRE_LINK_RESP,
			Sequence: seqID,
		},
	}
	data, _ := b.IEncode()
	return data
}

// NewUnBindRespBytes ...
func NewUnBindRespBytes(seqID uint32) []byte {
	b := &UnBindResp{
		Header: smpp.Header{
			ID:       smpp.UNBIND_RESP,
			Sequence: seqID,
		},
	}
	data, _ := b.IEncode()
	return data
}

// NewDeliverySMRespBytes ...
func NewDeliverySMRespBytes(seqID uint32) []byte {
	b := &DeliverSmResp{
		Header: smpp.Header{
			ID:       smpp.DELIVER_SM_RESP,
			Sequence: seqID,
		},
	}
	data, _ := b.IEncode()
	return data
}

// NewUnBindBytes ...
func NewUnBindBytes(seqID uint32) []byte {
	u := &Unbind{
		Header: smpp.Header{
			ID:       smpp.UNBIND,
			Sequence: seqID,
		},
	}
	data, _ := u.IEncode()
	return data
}

func DecodeSMPP34(data []byte) (sms.PDU, error) {
	header, err := smpp.PeekHeader(data)
	if err != nil {
		return nil, err
	}

	var pdu sms.PDU
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
		return nil, sms.ErrUnsupportedPacket
	}

	err = pdu.IDecode(data)
	if err != nil {
		return nil, err
	}

	return pdu, nil
}
