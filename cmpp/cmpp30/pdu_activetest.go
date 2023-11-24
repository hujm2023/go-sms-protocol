package cmpp30

import (
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/packet"
)

type ActiveTest struct {
	cmpp.Header
}

func (p *ActiveTest) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter()
	defer buf.Release()

	buf.WriteUint32(uint32(p.Header.CommandID))
	buf.WriteUint32(p.Header.SequenceID)

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

type ActiveTestResp struct {
	cmpp.Header

	// 1 字节，保留字段
	Reserved uint8
}

func (pr *ActiveTestResp) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter()
	defer buf.Release()

	buf.WriteUint32(uint32(pr.Header.CommandID))
	buf.WriteUint32(pr.Header.SequenceID)
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
