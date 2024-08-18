package smgp30

import (
	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/packet"
	"github.com/hujm2023/go-sms-protocol/smgp"
)

type ActiveTest struct {
	smgp.Header
}

func (p *ActiveTest) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter()
	defer buf.Release()

	smgp.WriteHeaderNoLength(p.Header, buf)

	return buf.BytesWithLength()
}

func (p *ActiveTest) IDecode(data []byte) error {
	if len(data) < smgp.MinSMGPPduLength {
		return smgp.ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	p.Header = smgp.ReadHeader(buf)

	return buf.Error()
}

func (p *ActiveTest) SetSequenceID(id uint32) {
	p.Header.SequenceID = id
}

func (p *ActiveTest) GetSequenceID() uint32 {
	return p.Header.SequenceID
}

func (p *ActiveTest) GenerateResponseHeader() *ActiveTestResp {
	resp := &ActiveTestResp{
		Header: smgp.NewHeader(smgp.MaxActiveTestRespLength, smgp.CommandActiveTestResp, p.GetSequenceID()),
	}
	return resp
}

func (a *ActiveTest) GetCommand() sms.ICommander {
	return smgp.CommandActiveTest
}

func (a *ActiveTest) GenEmptyResponse() sms.PDU {
	return &ActiveTestResp{
		Header: smgp.Header{
			CommandID:  smgp.CommandActiveTestResp,
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
	smgp.Header

	// 1 字节，保留字段
	Reserved uint8
}

func (pr *ActiveTestResp) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter()
	defer buf.Release()

	smgp.WriteHeaderNoLength(pr.Header, buf)
	buf.WriteUint8(pr.Reserved)

	return buf.BytesWithLength()
}

func (pr *ActiveTestResp) IDecode(data []byte) error {
	if len(data) < smgp.MinSMGPPduLength {
		return smgp.ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	pr.Header = smgp.ReadHeader(buf)
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
	return smgp.CommandActiveTestResp
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
