package smgp30

import (
	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/packet"
	"github.com/hujm2023/go-sms-protocol/smgp"
)

type Login struct {
	Header smgp.Header

	// 8字节，用户账号
	ClientID string

	// 16字节 认证码
	AuthenticatorClient string

	// 登陆类型
	LoginMode uint8

	// 1字节，版本号。高 4bit 表示主版本号，低 4bit 表示次版本号。对于 3.0 以下版本，固定 高4bit为3，低4位为0
	Version uint8

	// 4字节，时间戳，由客户端产生，格式为 MMDDHHMMSS
	Timestamp uint32
}

func (p *Login) IDecode(data []byte) error {
	if len(data) < smgp.MinSMGPPduLength {
		return smgp.ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	p.Header = smgp.ReadHeader(buf)
	p.ClientID = buf.ReadCStringN(8)
	p.AuthenticatorClient = buf.ReadCStringN(16)
	p.LoginMode = buf.ReadUint8()
	p.Timestamp = buf.ReadUint32()
	p.Version = buf.ReadUint8()

	return buf.Error()
}

func (p *Login) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter()
	defer buf.Release()

	smgp.WriteHeaderNoLength(p.Header, buf)
	buf.WriteFixedLenString(p.ClientID, 8)
	buf.WriteFixedLenString(p.AuthenticatorClient, 16)
	buf.WriteUint8(p.LoginMode)
	buf.WriteUint32(p.Timestamp)
	buf.WriteUint8(p.Version)

	return buf.BytesWithLength()
}

func (p *Login) SetSequenceID(id uint32) {
	p.Header.SequenceID = id
}

func (l *Login) GetSequenceID() uint32 {
	return l.Header.SequenceID
}

func (l *Login) GetCommand() sms.ICommander {
	return smgp.CommandLogin
}

func (l *Login) GenEmptyResponse() sms.PDU {
	return &LoginResp{
		Header: smgp.NewHeader(0, smgp.CommandLoginResp, l.GetSequenceID()),
	}
}

func (l *Login) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", l.Header)

	w.Write("ClientID", l.ClientID)
	w.WriteWithBytes("AuthenticatorClient", l.AuthenticatorClient)
	w.Write("LoginMode", l.LoginMode)
	w.Write("Timestamp", l.Timestamp)
	w.Write("Version", l.Version)

	return w.String()
}

type LoginResp struct {
	smgp.Header

	// 4 字节，状态：0正确 1消息结构错 2非法源地址 3认证错 4版本太高 >5其他错误
	Status LoginRespStatus

	// 16 字节
	AuthenticatorServer string

	// 1 字节，服务器支持的最高版本号
	ServerVersion uint8
}

func (c *LoginResp) IDecode(data []byte) error {
	if len(data) < smgp.MinSMGPPduLength {
		return smgp.ErrInvalidPudLength
	}
	buf := packet.NewPacketReader(data)
	defer buf.Release()

	c.Header = smgp.ReadHeader(buf)
	c.Status = LoginRespStatus(buf.ReadUint32())
	c.AuthenticatorServer = buf.ReadCStringN(16)
	c.ServerVersion = buf.ReadUint8()
	return buf.Error()
}

func (c *LoginResp) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter()
	defer buf.Release()

	smgp.WriteHeaderNoLength(c.Header, buf)
	buf.WriteUint32(uint32(c.Status))
	buf.WriteFixedLenString(c.AuthenticatorServer, 16)
	buf.WriteUint8(c.ServerVersion)

	return buf.BytesWithLength()
}

func (c *LoginResp) SetSequenceID(id uint32) {
	c.Header.SequenceID = id
}

func (l *LoginResp) GetSequenceID() uint32 {
	return l.Header.SequenceID
}

func (l *LoginResp) GetCommand() sms.ICommander {
	return smgp.CommandLoginResp
}

func (l *LoginResp) GenEmptyResponse() sms.PDU {
	return nil
}

func (l *LoginResp) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", l.Header)
	w.Write("Status", l.Status)
	w.WriteWithBytes("AuthenticatorServer", l.AuthenticatorServer)
	w.Write("ServerVersion", l.ServerVersion)

	return w.String()
}

// ------

type LoginRespStatus uint32

const (
	LoginRespStatusSuccess LoginRespStatus = iota
	LoginRespStatusMsgStructErr
	LoginRespStatusInvalidSourceAddr
	LoginRespStatusAuthErr
	LoginRespStatusVersionTooHigh
	LoginRespStatusOtherErr
)

func (l LoginRespStatus) String() string {
	switch l {
	case LoginRespStatusSuccess:
		return "Success"
	case LoginRespStatusMsgStructErr:
		return "MsgStructErr"
	case LoginRespStatusInvalidSourceAddr:
		return "InvalidSourceAddr"
	case LoginRespStatusAuthErr:
		return "AuthErr"
	case LoginRespStatusVersionTooHigh:
		return "VersionTooHigh"
	case LoginRespStatusOtherErr:
		return "OtherErr"
	default:
		return "Unknown"
	}
}
