package cmpp20

import (
	"github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/packet"
)

type PduTerminate struct {
	cmpp.Header
}

func (p *PduTerminate) IEncode() ([]byte, error) {
	p.TotalLength = MaxTerminateLength
	buf := packet.NewPacketWriter(MaxTerminateLength)
	defer buf.Release()

	// header
	buf.WriteBytes(p.Header.Bytes())

	return buf.Bytes()
}

func (p *PduTerminate) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	p.Header = cmpp.ReadHeader(buf)

	return buf.Error()
}

func (p *PduTerminate) GetHeader() cmpp.Header {
	return p.Header
}

func (p *PduTerminate) GetSequenceID() uint32 {
	return p.GetHeader().SequenceID
}

func (p *PduTerminate) GetCommandID() cmpp.CommandID {
	return cmpp.CommandTerminate
}

func (p *PduTerminate) NewHeader(seqID uint32) cmpp.Header {
	return cmpp.NewHeader(MaxTerminateLength, cmpp.CommandTerminate, seqID)
}

func (p *PduTerminate) GenerateResponseHeader() protocol.PDU {
	resp := &PduTerminateResp{
		Header: cmpp.NewHeader(MaxTerminateRespLength, cmpp.CommandTerminateResp, p.GetSequenceID()),
	}
	return resp
}

func (p *PduTerminate) SetSequenceID(id uint32) {
	p.Header.SequenceID = id
}

func (p *PduTerminate) MaxLength() uint32 {
	return MaxTerminateLength
}

// --------------

type PduTerminateResp struct {
	cmpp.Header
}

func (p *PduTerminateResp) IEncode() ([]byte, error) {
	p.TotalLength = MaxTerminateRespLength
	buf := packet.NewPacketWriter(MaxTerminateRespLength)
	defer buf.Release()

	// header
	buf.WriteBytes(p.Header.Bytes())

	return buf.Bytes()
}

func (p *PduTerminateResp) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	p.Header = cmpp.ReadHeader(buf)

	return buf.Error()
}

func (p *PduTerminateResp) GetHeader() cmpp.Header {
	return p.Header
}

func (p *PduTerminateResp) GetSequenceID() uint32 {
	return p.GetHeader().SequenceID
}

func (p *PduTerminateResp) GetCommandID() cmpp.CommandID {
	return cmpp.CommandTerminateResp
}

func (p *PduTerminateResp) NewHeader(seqID uint32) cmpp.Header {
	return cmpp.NewHeader(MaxTerminateRespLength, cmpp.CommandTerminateResp, seqID)
}

func (p *PduTerminateResp) GenerateResponseHeader() protocol.PDU {
	return nil
}

func (p *PduTerminateResp) SetSequenceID(id uint32) {
	p.Header.SequenceID = id
}

func (p *PduTerminateResp) MaxLength() uint32 {
	return MaxTerminateRespLength
}
