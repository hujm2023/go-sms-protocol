package cmpp30

import (
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/packet"
)

type Cancel struct {
	cmpp.Header

	// 8字节，信息标识（SP 想要删除的信息标识）
	MsgID uint64
}

func (c *Cancel) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return cmpp.ErrInvalidPudLength
	}

	b := packet.NewPacketReader(data)
	defer b.Release()

	c.Header = cmpp.ReadHeader(b)
	c.MsgID = b.ReadUint64()
	return b.Error()
}

func (c *Cancel) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()

	cmpp.WriteHeaderNoLength(c.Header, b)
	b.WriteUint64(c.MsgID)
	return b.BytesWithLength()
}

func (c *Cancel) SetSequenceID(id uint32) {
	c.Header.SequenceID = id
}

type CancelResp struct {
	Header cmpp.Header

	// 4 成功标识，0：成功 1：失败
	SuccessID uint32
}

func (c *CancelResp) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return cmpp.ErrInvalidPudLength
	}
	b := packet.NewPacketReader(data)
	defer b.Release()

	c.Header = cmpp.ReadHeader(b)
	c.SuccessID = b.ReadUint32()
	return b.Error()
}

func (c *CancelResp) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()

	cmpp.WriteHeaderNoLength(c.Header, b)
	b.WriteUint32(c.SuccessID)
	return b.BytesWithLength()
}

func (c *CancelResp) SetSequenceID(id uint32) {
	c.Header.SequenceID = id
}
