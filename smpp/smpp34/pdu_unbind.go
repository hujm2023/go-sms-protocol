package smpp34

import (
	"github.com/hujm2023/go-sms-protocol/packet"
)

type Unbind struct {
	Header
}

func (u *Unbind) IDecode(data []byte) error {
	if len(data) < MinSMPPPacketLen {
		return ErrInvalidPudLength
	}
	buf := packet.NewPacketReader(data)
	defer buf.Release()

	u.Header = ReadHeader(buf)
	return buf.Error()
}

func (u *Unbind) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	buf.WriteUint32(uint32(u.Header.ID))
	buf.WriteUint8(uint8(u.Header.Status))
	buf.WriteUint32(u.Header.Sequence)

	return buf.BytesWithLength()
}

func (u *Unbind) SetSequenceID(id uint32) {
	u.Header.Sequence = id
}

type UnBindResp struct {
	Header
}

func (u *UnBindResp) IDecode(data []byte) error {
	if len(data) < MinSMPPPacketLen {
		return ErrInvalidPudLength
	}
	buf := packet.NewPacketReader(data)
	defer buf.Release()

	u.Header = ReadHeader(buf)
	return buf.Error()
}

func (u *UnBindResp) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	buf.WriteUint32(uint32(u.Header.ID))
	buf.WriteUint8(uint8(u.Header.Status))
	buf.WriteUint32(u.Header.Sequence)

	return buf.BytesWithLength()
}

func (u *UnBindResp) SetSequenceID(id uint32) {
	u.Header.Sequence = id
}
