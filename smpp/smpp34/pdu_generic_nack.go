package smpp34

import (
	sms "github.com/hujm2023/go-sms-protocol"
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

func (g *GenericNack) GetSequenceID() uint32 {
	return g.Header.Sequence
}

func (g *GenericNack) GetCommand() sms.ICommander {
	return smpp.GENERIC_NACK
}

func (g *GenericNack) GenEmptyResponse() sms.PDU {
	return nil
}

func (g *GenericNack) String() string {
	s := packet.NewPDUStringer()
	defer s.Release()

	s.Write("Header", g.Header)

	return s.String()
}
