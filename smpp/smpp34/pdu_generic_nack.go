package smpp34

import (
	"github.com/hujm2023/go-sms-protocol/packet"
)

type GenericNack struct {
	Header
}

func (g *GenericNack) IDecode(data []byte) error {
	if len(data) < MinSMPPPacketLen {
		return ErrInvalidPudLength
	}
	buf := packet.NewPacketReader(data)
	defer buf.Release()

	g.Header = ReadHeader(buf)
	return nil
}

func (g *GenericNack) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	buf.WriteUint32(uint32(g.Header.ID))
	buf.WriteUint32(uint32(g.Header.Status))
	buf.WriteUint32(g.Header.Sequence)

	return buf.BytesWithLength()
}

func (g *GenericNack) SetSequenceID(id uint32) {
	g.Header.Sequence = id
}
