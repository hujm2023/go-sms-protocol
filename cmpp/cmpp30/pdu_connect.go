package cmpp30

import (
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/packet"
)

type Connect struct {
	cmpp.Header
	// 6字节，cmpp2 账号
	SourceAddr string

	// 16字节，认证，= md5(SourceAddr + 9字节的0 + 密码 + Timestamp)
	AuthenticatorSource string

	// 1字节，版本号。高 4bit 表示主版本号，低 4bit 表示次版本号。对于 3.0 以下版本，固定 高4bit为3，低4位为0
	Version uint8

	// 4字节，时间戳，由客户端产生，格式为 MMDDHHMMSS
	Timestamp uint32
}

func (p *Connect) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return cmpp.ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	p.Header = cmpp.ReadHeader(buf)
	p.SourceAddr = buf.ReadCStringN(6)
	p.AuthenticatorSource = buf.ReadCStringN(16)
	p.Version = buf.ReadUint8()
	p.Timestamp = buf.ReadUint32()

	return buf.Error()
}

func (p *Connect) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter()
	defer buf.Release()

	cmpp.WriteHeaderNoLength(p.Header, buf)
	buf.WriteFixedLenString(p.SourceAddr, 6)
	buf.WriteFixedLenString(p.AuthenticatorSource, 16)
	buf.WriteUint8(p.Version)
	buf.WriteUint32(p.Timestamp)

	return buf.BytesWithLength()
}

func (p *Connect) SetSequenceID(id uint32) {
	p.Header.SequenceID = id
}

type ConnectResp struct {
	cmpp.Header

	// 4 字节，状态：0正确 1消息结构错 2非法源地址 3认证错 4版本太高 >5其他错误
	Status uint32

	// 16 字节，ISMG 认证码，用于鉴别 ISMG， = md5(Status + req.AuthenticatorSource + 密码)
	AuthenticatorISMG string

	// 1 字节，服务器支持的最高版本号
	Version uint8
}

func (c *ConnectResp) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return cmpp.ErrInvalidPudLength
	}
	buf := packet.NewPacketReader(data)
	defer buf.Release()

	c.Header = cmpp.ReadHeader(buf)
	c.Status = buf.ReadUint32()
	c.AuthenticatorISMG = buf.ReadCStringN(16)
	c.Version = buf.ReadUint8()
	return buf.Error()
}

func (c *ConnectResp) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter()
	defer buf.Release()

	cmpp.WriteHeaderNoLength(c.Header, buf)
	buf.WriteUint32(c.Status)
	buf.WriteFixedLenString(c.AuthenticatorISMG, 16)
	buf.WriteUint8(c.Version)

	return buf.BytesWithLength()
}

func (c *ConnectResp) SetSequenceID(id uint32) {
	c.Header.SequenceID = id
}
