package smpp50

import (
	"github.com/hujm2023/go-sms-protocol/packet"
	"github.com/hujm2023/go-sms-protocol/smpp"
)

type Unbind struct {
	smpp.Header

	// CString, max 16
	SystemID string

	// CString, max 9
	Password string
}

func (u *Unbind) IDecode(data []byte) error {
	buf := packet.NewPacketReader(data)
	defer buf.Release()

	u.Header = smpp.ReadHeader(buf)
	u.SystemID = buf.ReadCString()
	u.Password = buf.ReadCString()

	return buf.Error()
}

func (u *Unbind) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter()
	defer buf.Release()

	smpp.WriteHeaderNoLength(u.Header, buf)
	buf.WriteCString(u.SystemID)
	buf.WriteCString(u.Password)

	return buf.BytesWithLength()
}

func (u *Unbind) SetSequenceID(id uint32) {
	u.Header.Sequence = id
}

type UnbindResp struct {
	smpp.Header
}

func (u *UnbindResp) IDecode(data []byte) error {
	buf := packet.NewPacketReader(data)
	defer buf.Release()

	u.Header = smpp.ReadHeader(buf)

	return buf.Error()
}

func (u *UnbindResp) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter()
	defer buf.Release()

	smpp.WriteHeaderNoLength(u.Header, buf)

	return buf.BytesWithLength()
}

func (u *UnbindResp) SetSequenceID(id uint32) {
	u.Header.Sequence = id
}
