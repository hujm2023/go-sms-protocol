package smpp34

import (
	"fmt"

	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/packet"
	"github.com/hujm2023/go-sms-protocol/smpp"
)

type EnquireLink struct {
	smpp.Header
}

func (e *EnquireLink) IDecode(data []byte) error {
	header, err := smpp.ValidateDecodedPDU(data, smpp.ENQUIRE_LINK, false)
	if err != nil {
		return err
	}
	if len(data) != smpp.MinSMPPPacketLen {
		return fmt.Errorf("enquire_link must not contain a body")
	}
	if err := smpp.ValidateRequestHeader(header, smpp.ENQUIRE_LINK); err != nil {
		return err
	}
	e.Header = header
	return nil
}

func (e *EnquireLink) IEncode() ([]byte, error) {
	if err := smpp.ValidateRequestHeader(e.Header, smpp.ENQUIRE_LINK); err != nil {
		return nil, err
	}
	buf := packet.NewPacketWriter(0)
	defer buf.Release()
	smpp.WriteHeaderNoLength(e.Header, buf)
	data, err := buf.BytesWithLength()
	if err != nil {
		return nil, err
	}
	return data, smpp.ValidateEncodedPDU(data)
}

func (e *EnquireLink) SetSequenceID(id uint32) {
	e.Header.Sequence = id
}

func (e *EnquireLink) GetSequenceID() uint32 {
	return e.Header.Sequence
}

func (e *EnquireLink) GetCommand() sms.ICommander {
	return smpp.ENQUIRE_LINK
}

func (e *EnquireLink) GenEmptyResponse() sms.PDU {
	return &EnquireLinkResp{
		Header: smpp.Header{
			ID:       smpp.ENQUIRE_LINK_RESP,
			Sequence: e.Header.Sequence,
		},
	}
}

func (e *EnquireLink) String() string {
	s := packet.NewPDUStringer()
	defer s.Release()

	s.Write("Header", e.Header)

	return s.String()
}

type EnquireLinkResp struct {
	smpp.Header
}

func (e *EnquireLinkResp) GetSequenceID() uint32 {
	return e.Header.Sequence
}

func (e *EnquireLinkResp) GetCommand() sms.ICommander {
	return smpp.ENQUIRE_LINK_RESP
}

func (e *EnquireLinkResp) GenEmptyResponse() sms.PDU {
	return nil
}

func (e *EnquireLinkResp) IDecode(data []byte) error {
	header, err := smpp.ValidateDecodedPDU(data, smpp.ENQUIRE_LINK_RESP, false)
	if err != nil {
		return err
	}
	if len(data) != smpp.MinSMPPPacketLen {
		return fmt.Errorf("enquire_link_resp must not contain a body")
	}
	if err := smpp.ValidateResponseHeader(header, smpp.ENQUIRE_LINK_RESP); err != nil {
		return err
	}
	e.Header = header
	return nil
}

func (e *EnquireLinkResp) IEncode() ([]byte, error) {
	if err := smpp.ValidateResponseHeader(e.Header, smpp.ENQUIRE_LINK_RESP); err != nil {
		return nil, err
	}
	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	smpp.WriteHeaderNoLength(e.Header, buf)

	data, err := buf.BytesWithLength()
	if err != nil {
		return nil, err
	}
	return data, smpp.ValidateEncodedPDU(data)
}

func (e *EnquireLinkResp) SetSequenceID(id uint32) {
	e.Header.Sequence = id
}

func (e *EnquireLinkResp) String() string {
	s := packet.NewPDUStringer()
	defer s.Release()

	s.Write("Header", e.Header)

	return s.String()
}
