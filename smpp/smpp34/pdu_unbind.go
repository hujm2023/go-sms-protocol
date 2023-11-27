package smpp34

import (
	"github.com/hujm2023/go-sms-protocol/packet"
	"github.com/hujm2023/go-sms-protocol/smpp"
)

type Unbind struct {
	smpp.Header
}

func (u *Unbind) IDecode(data []byte) error {
	if len(data) < smpp.MinSMPPPacketLen {
		return smpp.ErrInvalidPudLength
	}
	buf := packet.NewPacketReader(data)
	defer buf.Release()

	u.Header = smpp.ReadHeader(buf)
	return buf.Error()
}

func (u *Unbind) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	smpp.WriteHeaderNoLength(u.Header, buf)

	return buf.BytesWithLength()
}

func (u *Unbind) SetSequenceID(id uint32) {
	u.Header.Sequence = id
}

type UnBindResp struct {
	smpp.Header
}

func (u *UnBindResp) IDecode(data []byte) error {
	if len(data) < smpp.MinSMPPPacketLen {
		return smpp.ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	u.Header = smpp.ReadHeader(buf)
	return buf.Error()
}

func (u *UnBindResp) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	smpp.WriteHeaderNoLength(u.Header, buf)

	return buf.BytesWithLength()
}

func (u *UnBindResp) SetSequenceID(id uint32) {
	u.Header.Sequence = id
}
