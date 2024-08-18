package smpp34

import (
	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/packet"
	"github.com/hujm2023/go-sms-protocol/smpp"
)

type EnquireLink struct {
	smpp.Header
}

func (e *EnquireLink) IDecode(data []byte) error {
	if len(data) < smpp.MinSMPPPacketLen {
		return smpp.ErrInvalidPudLength
	}
	buf := packet.NewPacketReader(data)
	defer buf.Release()

	e.Header = smpp.ReadHeader(buf)
	return buf.Error()
}

func (e *EnquireLink) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter(0)
	smpp.WriteHeaderNoLength(e.Header, buf)
	return buf.BytesWithLength()
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
	if len(data) < smpp.MinSMPPPacketLen {
		return smpp.ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	e.Header = smpp.ReadHeader(buf)
	return buf.Error()
}

func (e *EnquireLinkResp) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	smpp.WriteHeaderNoLength(e.Header, buf)

	return buf.BytesWithLength()
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
