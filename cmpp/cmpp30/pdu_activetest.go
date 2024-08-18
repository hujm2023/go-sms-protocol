package cmpp30

import (
	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/packet"
)

type ActiveTest struct {
	cmpp.Header
}

func (p *ActiveTest) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter()
	defer buf.Release()

	cmpp.WriteHeaderNoLength(p.Header, buf)

	return buf.BytesWithLength()
}

func (p *ActiveTest) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return cmpp.ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	p.Header = cmpp.ReadHeader(buf)

	return buf.Error()
}

func (p *ActiveTest) SetSequenceID(id uint32) {
	p.Header.SequenceID = id
}

func (a *ActiveTest) GetSequenceID() uint32 {
	return a.Header.SequenceID
}

func (a *ActiveTest) GetCommand() sms.ICommander {
	return cmpp.CommandActiveTest
}

func (a *ActiveTest) GenEmptyResponse() sms.PDU {
	return &ActiveTestResp{
		Header: cmpp.Header{
			CommandID:  cmpp.CommandActiveTestResp,
			SequenceID: a.GetSequenceID(),
		},
	}
}

func (a *ActiveTest) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", a.Header)

	return w.String()
}

type ActiveTestResp struct {
	cmpp.Header

	// 1 字节，保留字段
	Reserved uint8
}

func (pr *ActiveTestResp) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter()
	defer buf.Release()

	cmpp.WriteHeaderNoLength(pr.Header, buf)
	buf.WriteUint8(pr.Reserved)

	return buf.BytesWithLength()
}

func (pr *ActiveTestResp) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return cmpp.ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	pr.Header = cmpp.ReadHeader(buf)
	pr.Reserved = buf.ReadUint8()

	return buf.Error()
}

func (pr *ActiveTestResp) SetSequenceID(sid uint32) {
	pr.Header.SequenceID = sid
}

func (a *ActiveTestResp) GetSequenceID() uint32 {
	return a.Header.SequenceID
}

func (a *ActiveTestResp) GetCommand() sms.ICommander {
	return cmpp.CommandActiveTestResp
}

func (a *ActiveTestResp) GenEmptyResponse() sms.PDU {
	return nil
}

func (a *ActiveTestResp) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", a.Header)
	w.Write("Reserved", a.Reserved)

	return w.String()
}
