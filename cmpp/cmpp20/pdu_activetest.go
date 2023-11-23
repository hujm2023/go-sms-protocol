package cmpp20

import (
	protocol "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/packet"
)

type PduActiveTest struct {
	cmpp.Header
}

func (p *PduActiveTest) IEncode() ([]byte, error) {
	p.TotalLength = MaxActiveTestLength
	buf := packet.NewPacketWriter(MaxActiveTestLength)
	defer buf.Release()

	buf.WriteBytes(p.Header.Bytes())

	return buf.Bytes()
}

func (p *PduActiveTest) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	p.Header = cmpp.ReadHeader(buf)

	return buf.Error()
}

func (p *PduActiveTest) GetHeader() cmpp.Header {
	return p.Header
}

func (p *PduActiveTest) GetSequenceID() uint32 {
	return p.GetHeader().SequenceID
}

func (p *PduActiveTest) GetCommandID() cmpp.CommandID {
	return cmpp.CommandActiveTest
}

func (p *PduActiveTest) GenerateResponseHeader() protocol.PDU {
	resp := &PduActiveTestResp{
		Header: cmpp.NewHeader(MaxActiveTestRespLength, cmpp.CommandActiveTestResp, p.GetSequenceID()),
	}
	return resp
}

func (p *PduActiveTest) MaxLength() uint32 {
	return MaxActiveTestLength
}

func (p *PduActiveTest) SetSequenceID(sid uint32) {
	p.Header.SequenceID = sid
}

// --------------------------------------------------------------------

type PduActiveTestResp struct {
	cmpp.Header

	// 1 字节，保留字段
	Reserved uint8
}

func (pr *PduActiveTestResp) IEncode() ([]byte, error) {
	pr.TotalLength = MaxActiveTestRespLength
	buf := packet.NewPacketWriter(MaxActiveTestRespLength)
	defer buf.Release()

	buf.WriteBytes(pr.Header.Bytes())
	buf.WriteUint8(pr.Reserved)

	return buf.Bytes()
}

func (pr *PduActiveTestResp) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	pr.Header = cmpp.ReadHeader(buf)
	pr.Reserved = buf.ReadUint8()

	return buf.Error()
}

func (pr *PduActiveTestResp) GetHeader() cmpp.Header {
	return pr.Header
}

func (pr *PduActiveTestResp) GetSequenceID() uint32 {
	return pr.GetHeader().SequenceID
}

func (pr *PduActiveTestResp) GetCommandID() cmpp.CommandID {
	return cmpp.CommandActiveTestResp
}

func (pr *PduActiveTestResp) GenerateResponseHeader() protocol.PDU {
	return nil
}

func (pr *PduActiveTestResp) MaxLength() uint32 {
	return MaxActiveTestRespLength
}

func (pr *PduActiveTestResp) SetSequenceID(sid uint32) {
	pr.Header.SequenceID = sid
}
