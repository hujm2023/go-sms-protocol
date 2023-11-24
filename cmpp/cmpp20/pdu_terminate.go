package cmpp20

import (
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/packet"
)

type PduTerminate struct {
	cmpp.Header
}

func (p *PduTerminate) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter()
	defer buf.Release()

	buf.WriteUint32(uint32(p.Header.CommandID))
	buf.WriteUint32(p.Header.SequenceID)

	return buf.BytesWithLength()
}

func (p *PduTerminate) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return cmpp.ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	p.Header = cmpp.ReadHeader(buf)

	return buf.Error()
}

func (p *PduTerminate) SetSequenceID(id uint32) {
	p.Header.SequenceID = id
}

// --------------

type PduTerminateResp struct {
	cmpp.Header
}

func (p *PduTerminateResp) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter()
	defer buf.Release()

	// header
	buf.WriteUint32(uint32(p.Header.CommandID))
	buf.WriteUint32(p.Header.SequenceID)

	return buf.BytesWithLength()
}

func (p *PduTerminateResp) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return cmpp.ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	p.Header = cmpp.ReadHeader(buf)

	return buf.Error()
}

func (p *PduTerminateResp) SetSequenceID(id uint32) {
	p.Header.SequenceID = id
}
