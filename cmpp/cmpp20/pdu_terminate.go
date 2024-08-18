package cmpp20

import (
	sms "github.com/hujm2023/go-sms-protocol"
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

func (p *PduTerminate) GetSequenceID() uint32 {
	return p.Header.SequenceID
}

func (p *PduTerminate) SetSequenceID(id uint32) {
	p.Header.SequenceID = id
}

func (p *PduTerminate) GetCommand() sms.ICommander {
	return cmpp.CommandTerminate
}

func (p *PduTerminate) GenEmptyResponse() sms.PDU {
	return &PduTerminateResp{
		Header: cmpp.Header{
			CommandID:  cmpp.CommandTerminateResp,
			SequenceID: p.GetSequenceID(),
		},
	}
}

func (p *PduTerminate) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", p.Header)

	return w.String()
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

func (p *PduTerminateResp) GetSequenceID() uint32 {
	return p.Header.SequenceID
}

func (p *PduTerminateResp) SetSequenceID(id uint32) {
	p.Header.SequenceID = id
}

func (p *PduTerminateResp) GetCommand() sms.ICommander {
	return cmpp.CommandTerminateResp
}

func (p *PduTerminateResp) GenEmptyResponse() sms.PDU {
	return nil
}

func (p *PduTerminateResp) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", p.Header)

	return w.String()
}
