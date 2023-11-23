package smpp34

import (
	"github.com/hujm2023/go-sms-protocol/packet"
)

type EnquireLink struct {
	Header
}

func (e *EnquireLink) IDecode(data []byte) error {
	if len(data) < MinSMPPPacketLen {
		return ErrInvalidPudLength
	}
	buf := packet.NewPacketReader(data)
	defer buf.Release()

	e.Header = ReadHeader(buf)
	return buf.Error()
}

func (e *EnquireLink) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter(0)
	buf.WriteUint32(uint32(e.Header.ID))
	buf.WriteUint32(uint32(e.Header.Status))
	buf.WriteUint32(e.Header.Sequence)
	return buf.BytesWithLength()
}

func (e *EnquireLink) SetSequenceID(id uint32) {
	e.Header.Sequence = id
}

type EnquireLinkResp struct {
	Header
}

func (e *EnquireLinkResp) IDecode(data []byte) error {
	if len(data) < MinSMPPPacketLen {
		return ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	e.Header = ReadHeader(buf)
	return buf.Error()
}

func (e *EnquireLinkResp) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter(0)
	buf.WriteUint32(uint32(e.Header.ID))
	buf.WriteUint32(uint32(e.Header.Status))
	buf.WriteUint32(e.Header.Sequence)
	return buf.BytesWithLength()
}

func (e *EnquireLinkResp) SetSequenceID(id uint32) {
	e.Header.Sequence = id
}
