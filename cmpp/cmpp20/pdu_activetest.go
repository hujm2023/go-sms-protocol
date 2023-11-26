package cmpp20

import (
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/packet"
)

type PduActiveTest struct {
	cmpp.Header
}

func (p *PduActiveTest) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter()
	defer buf.Release()

	cmpp.WriteHeaderNoLength(p.Header, buf)

	return buf.BytesWithLength()
}

func (p *PduActiveTest) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return cmpp.ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	p.Header = cmpp.ReadHeader(buf)

	return buf.Error()
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
	buf := packet.NewPacketWriter()
	defer buf.Release()

	cmpp.WriteHeaderNoLength(pr.Header, buf)
	buf.WriteUint8(pr.Reserved)

	return buf.BytesWithLength()
}

func (pr *PduActiveTestResp) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return cmpp.ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	pr.Header = cmpp.ReadHeader(buf)
	pr.Reserved = buf.ReadUint8()

	return buf.Error()
}

func (pr *PduActiveTestResp) SetSequenceID(sid uint32) {
	pr.Header.SequenceID = sid
}
