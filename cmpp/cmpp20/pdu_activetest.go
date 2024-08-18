package cmpp20

import (
	sms "github.com/hujm2023/go-sms-protocol"
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

func (p *PduActiveTest) GetSequenceID() uint32 {
	return p.Header.SequenceID
}

func (p *PduActiveTest) SetSequenceID(sid uint32) {
	p.Header.SequenceID = sid
}

func (p *PduActiveTest) GetCommand() sms.ICommander {
	return cmpp.CommandActiveTest
}

func (p *PduActiveTest) GenEmptyResponse() sms.PDU {
	return &PduActiveTestResp{
		Header: cmpp.Header{
			CommandID:  cmpp.CommandActiveTestResp,
			SequenceID: p.GetSequenceID(),
		},
	}
}

func (p *PduActiveTest) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", p.Header)

	return w.String()
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

func (pr *PduActiveTestResp) GetSequenceID() uint32 {
	return pr.Header.SequenceID
}

func (pr *PduActiveTestResp) SetSequenceID(sid uint32) {
	pr.Header.SequenceID = sid
}

func (p *PduActiveTestResp) GetCommand() sms.ICommander {
	return cmpp.CommandActiveTestResp
}

func (p *PduActiveTestResp) GenEmptyResponse() sms.PDU {
	return nil
}

func (p *PduActiveTestResp) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", p.Header)
	w.Write("Reserved", p.Reserved)

	return w.String()
}
