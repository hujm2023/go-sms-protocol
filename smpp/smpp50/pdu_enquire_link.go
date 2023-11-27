package smpp50

import (
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
	buf := packet.NewPacketWriter()
	defer buf.Release()

	smpp.WriteHeaderNoLength(e.Header, buf)

	return buf.BytesWithLength()
}

func (e *EnquireLink) SetSequenceID(id uint32) {
	e.Header.Sequence = id
}

type EnquireLinkResp struct {
	smpp.Header
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
	buf := packet.NewPacketWriter()
	defer buf.Release()

	smpp.WriteHeaderNoLength(e.Header, buf)

	return buf.BytesWithLength()
}

func (e *EnquireLinkResp) SetSequenceID(id uint32) {
	e.Header.Sequence = id
}
