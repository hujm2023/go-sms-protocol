package cmpp20

import (
	sms "github.com/hujm2023/go-sms-protocol"
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
	p.TotalLength = MaxConnectLength
	buf := packet.NewPacketWriter(MaxConnectLength)
	defer buf.Release()

	buf.WriteBytes(p.Header.Bytes())
	buf.WriteFixedLenString(p.SourceAddr, 6)
	buf.WriteFixedLenString(p.AuthenticatorSource, 16)
	buf.WriteUint8(p.Version)
	buf.WriteUint32(p.Timestamp)

	return buf.Bytes()
}

func (p *PduConnect) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return ErrInvalidPudLength
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

func (p *PduConnect) GetSequenceID() uint32 {
	return p.Header.SequenceID
}

func (p *PduConnect) SetSequenceID(sid uint32) {
	p.Header.SequenceID = sid
}

func (p *PduConnect) GetCommand() sms.ICommander {
	return cmpp.CommandConnect
}

func (p *PduConnect) GenEmptyResponse() sms.PDU {
	return &PduConnectResp{
		Header: cmpp.Header{
			CommandID:  cmpp.CommandConnectResp,
			SequenceID: p.GetSequenceID(),
		},
	}
}

func (p *PduConnect) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", p.Header)
	w.Write("SourceAddr", p.SourceAddr)
	w.WriteWithBytes("AuthenticatorSource", p.AuthenticatorSource)
	w.Write("Version", p.Version)
	w.Write("Timestamp", p.Timestamp)

	return w.String()
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
	pr.TotalLength = MaxConnectRespLength
	buf := packet.NewPacketWriter(MaxConnectRespLength)
	defer buf.Release()

	buf.WriteBytes(pr.Header.Bytes())
	buf.WriteUint8(pr.Status)
	buf.WriteFixedLenString(pr.AuthenticatorISMG, 16)
	buf.WriteUint8(pr.Version)

	return buf.Bytes()
}

func (pr *PduConnectResp) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	pr.Header = cmpp.ReadHeader(buf)
	pr.Status = buf.ReadUint8()
	pr.AuthenticatorISMG = buf.ReadCStringN(16)
	pr.Version = buf.ReadUint8()

	return buf.Error()
}

func (pr *PduConnectResp) GetSequenceID() uint32 {
	return pr.Header.SequenceID
}

func (pr *PduConnectResp) SetSequenceID(sid uint32) {
	pr.Header.SequenceID = sid
}

func (p *PduConnectResp) GetCommand() sms.ICommander {
	return cmpp.CommandConnectResp
}

func (p *PduConnectResp) GenEmptyResponse() sms.PDU {
	return nil
}

func (p *PduConnectResp) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", p.Header)
	w.Write("Status", p.Status)
	w.WriteWithBytes("AuthenticatorISMG", p.AuthenticatorISMG)
	w.Write("Version", p.Version)

	return w.String()
}
