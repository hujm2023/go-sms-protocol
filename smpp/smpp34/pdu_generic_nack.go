package smpp34

import (
	"github.com/hujm2023/go-sms-protocol/packet"
	"github.com/hujm2023/go-sms-protocol/smpp"
)

type GenericNack struct {
	smpp.Header
}

func (g *GenericNack) IDecode(data []byte) error {
	if len(data) < smpp.MinSMPPPacketLen {
		return smpp.ErrInvalidPudLength
	}
	buf := packet.NewPacketReader(data)
	defer buf.Release()

	g.Header = smpp.ReadHeader(buf)
	return nil
}

func (g *GenericNack) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	smpp.WriteHeaderNoLength(g.Header, buf)

	return buf.BytesWithLength()
}

func (g *GenericNack) SetSequenceID(id uint32) {
	g.Header.Sequence = id
}
