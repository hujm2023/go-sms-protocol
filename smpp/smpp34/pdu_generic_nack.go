package smpp34

import (
	"fmt"

	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/packet"
	"github.com/hujm2023/go-sms-protocol/smpp"
)

type GenericNack struct {
	smpp.Header
}

func (g *GenericNack) IDecode(data []byte) error {
	header, err := smpp.ValidateDecodedPDU(data, smpp.GENERIC_NACK, true)
	if err != nil {
		return err
	}
	if len(data) != smpp.MinSMPPPacketLen {
		return fmt.Errorf("generic_nack must not contain a body")
	}
	if err := smpp.ValidateGenericNackHeader(header); err != nil {
		return err
	}
	g.Header = header
	return nil
}

func (g *GenericNack) IEncode() ([]byte, error) {
	if err := smpp.ValidateGenericNackHeader(g.Header); err != nil {
		return nil, err
	}
	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	smpp.WriteHeaderNoLength(g.Header, buf)

	data, err := buf.BytesWithLength()
	if err != nil {
		return nil, err
	}
	return data, smpp.ValidateEncodedPDU(data)
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
