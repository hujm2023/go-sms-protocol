package smpp50

import (
	"github.com/hujm2023/go-sms-protocol/packet"
	"github.com/hujm2023/go-sms-protocol/smpp"
)

type GenericNack struct {
	smpp.Header
}

func (g *GenericNack) IDecode(data []byte) error {
	buf := packet.NewPacketReader(data)
	defer buf.Release()

	g.Header = smpp.ReadHeader(buf)

	return buf.Error()
}

func (g *GenericNack) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter()
	defer buf.Release()

	smpp.WriteHeaderNoLength(g.Header, buf)

	return buf.BytesWithLength()
}

func (g *GenericNack) SetSequenceID(id uint32) {
	g.Header.Sequence = id
}
