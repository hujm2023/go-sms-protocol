package smpp34

import (
	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/packet"
	"github.com/hujm2023/go-sms-protocol/smpp"
)

type Unbind struct {
	smpp.Header
}

func (u *Unbind) IDecode(data []byte) error {
	if len(data) < smpp.MinSMPPPacketLen {
		return smpp.ErrInvalidPudLength
	}
	buf := packet.NewPacketReader(data)
	defer buf.Release()

	u.Header = smpp.ReadHeader(buf)
	return buf.Error()
}

func (u *Unbind) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	smpp.WriteHeaderNoLength(u.Header, buf)

	return buf.BytesWithLength()
}

func (u *Unbind) SetSequenceID(id uint32) {
	u.Header.Sequence = id
}

func (u *Unbind) GetSequenceID() uint32 {
	return u.Header.Sequence
}

func (u *Unbind) GetCommand() sms.ICommander {
	return smpp.UNBIND
}

func (u *Unbind) GenEmptyResponse() sms.PDU {
	return &UnBindResp{
		Header: smpp.Header{
			ID:       smpp.UNBIND_RESP,
			Sequence: u.Header.Sequence,
		},
	}
}

type UnBindResp struct {
	smpp.Header
}

func (u *UnBindResp) IDecode(data []byte) error {
	if len(data) < smpp.MinSMPPPacketLen {
		return smpp.ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	u.Header = smpp.ReadHeader(buf)
	return buf.Error()
}

func (u *UnBindResp) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	smpp.WriteHeaderNoLength(u.Header, buf)

	return buf.BytesWithLength()
}

func (u *UnBindResp) SetSequenceID(id uint32) {
	u.Header.Sequence = id
}

func (u *UnBindResp) GetSequenceID() uint32 {
	return u.Header.Sequence
}

func (u *UnBindResp) GetCommand() sms.ICommander {
	return smpp.UNBIND_RESP
}

func (u *UnBindResp) GenEmptyResponse() sms.PDU {
	return nil
}

func (u *UnBindResp) String() string {
	str := packet.NewPDUStringer()
	defer str.Release()

	str.Write("Header", u.Header)

	return str.String()
}
