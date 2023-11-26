package cmpp20

import (
	"fmt"

	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/packet"
)

type PduConnect struct {
	cmpp.Header

	// 6字节，cmpp2 账号
	SourceAddr string

	// 16字节，认证，= md5(SourceAddr + 9字节的0 + 密码 + Timestamp)
	AuthenticatorSource string

	// 1 字节，版本号。高 4bit 表示主版本号，低 4bit 表示次版本号。对于 3.0 以下版本，固定 高4bit为3，低4位为0
	Version uint8

	// 4 字节，时间戳，由客户端产生，格式为 MMDDHHMMSS
	Timestamp uint32
}

func (p *PduConnect) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter()
	defer buf.Release()

	cmpp.WriteHeaderNoLength(p.Header, buf)
	buf.WriteFixedLenString(p.SourceAddr, 6)
	buf.WriteFixedLenString(p.AuthenticatorSource, 16)
	buf.WriteUint8(p.Version)
	buf.WriteUint32(p.Timestamp)

	return buf.BytesWithLength()
}

func (p *PduConnect) IDecode(data []byte) error {
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

func (p *PduConnect) SetSequenceID(sid uint32) {
	p.Header.SequenceID = sid
}

// --------------------------------------------------------------------

type PduConnectResp struct {
	cmpp.Header

	// 1 字节，状态：0正确 1消息结构错 2非法源地址 3认证错 4版本太高 >5其他错误
	Status uint8

	// 16 字节，ISMG 认证码，用于鉴别 ISMG， = md5(Status + req.AuthenticatorSource + 密码)
	AuthenticatorISMG string

	// 1 字节，服务器支持的最高版本号
	Version uint8
}

func (pr *PduConnectResp) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter()
	defer buf.Release()

	cmpp.WriteHeaderNoLength(pr.Header, buf)
	buf.WriteUint8(pr.Status)
	buf.WriteFixedLenString(pr.AuthenticatorISMG, 16)
	buf.WriteUint8(pr.Version)

	return buf.BytesWithLength()
}

func (pr *PduConnectResp) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return cmpp.ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	pr.Header = cmpp.ReadHeader(buf)
	pr.Status = buf.ReadUint8()
	pr.AuthenticatorISMG = buf.ReadCStringN(16)
	pr.Version = buf.ReadUint8()

	return buf.Error()
}

func (pr *PduConnectResp) SetSequenceID(sid uint32) {
	pr.Header.SequenceID = sid
}

// 1 字节，状态：0正确 1消息结构错 2非法源地址 3认证错 4版本太高 >5其他错误
var connectRespStatus = map[uint8]string{
	0: "正确",
	1: "消息结构错",
	2: "非法源地址",
	3: "认证错",
	4: "版本太高",
	5: "其他错误",
}

func ConnectRespResultString(r uint8) string {
	if s, ok := connectRespStatus[r]; ok {
		return s
	}
	return fmt.Sprintf("未知错误代码 %d", r)
}
