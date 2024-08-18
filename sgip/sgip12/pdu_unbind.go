package sgip12

import (
	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/packet"
	"github.com/hujm2023/go-sms-protocol/sgip"
)

// Unbind operation
type Unbind struct {
	sgip.Header
}

func (p *Unbind) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()
	sgip.WriteHeaderNoLength(p.Header, b)
	return b.BytesWithLength()
}

func (p *Unbind) IDecode(data []byte) error {
	if len(data) < sgip.MinSGIPPduLength {
		return sgip.ErrInvalidPudLength
	}
	b := packet.NewPacketReader(data)
	defer b.Release()
	p.Header = sgip.ReadHeader(b)
	return nil
}

func (p *Unbind) SetSequenceID(id uint32) {
	p.Header.Sequence[2] = id
}

func (u *Unbind) GetSequenceID() uint32 {
	return u.Header.Sequence[2]
}

func (u *Unbind) GetCommand() sms.ICommander {
	return sgip.SGIP_UNBIND
}

func (u *Unbind) GenEmptyResponse() sms.PDU {
	return &UnbindResp{
		Header: sgip.NewHeader(0, sgip.SGIP_UNBIND_REP, u.Sequence[0], u.GetSequenceID()),
	}
}

func (u *Unbind) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", u.Header)

	return w.String()
}

// ------------------------------------------------------------------------

// UnbindResp ...
type UnbindResp struct {
	sgip.Header
}

func (p *UnbindResp) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()
	sgip.WriteHeaderNoLength(p.Header, b)
	return b.BytesWithLength()
}

func (p *UnbindResp) IDecode(data []byte) error {
	if len(data) < sgip.MinSGIPPduLength {
		return sgip.ErrInvalidPudLength
	}
	b := packet.NewPacketReader(data)
	defer b.Release()
	p.Header = sgip.ReadHeader(b)
	return nil
}

func (p *UnbindResp) SetSequenceID(id uint32) {
	p.Header.Sequence[2] = id
}

func (u *UnbindResp) GetSequenceID() uint32 {
	return u.Header.Sequence[2]
}

func (u *UnbindResp) GetCommand() sms.ICommander {
	return sgip.SGIP_UNBIND_REP
}

func (u *UnbindResp) GenEmptyResponse() sms.PDU {
	return nil
}

func (u *UnbindResp) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", u.Header)

	return w.String()
}
