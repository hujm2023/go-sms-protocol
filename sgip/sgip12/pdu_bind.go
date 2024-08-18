package sgip12

import (
	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/packet"
	"github.com/hujm2023/go-sms-protocol/sgip"
)

type Bind struct {
	sgip.Header

	// 包含头部，总长20 + 1 + 16 + 16 + 8 = 61
	// 1 字节 登录类型
	Type sgip.LoginType

	// 16 字节 登录名
	Name string

	// 16 字节 密码
	Password string

	// 8  字节 保留
	Reserved string
}

func (p *Bind) IDecode(data []byte) error {
	if len(data) < sgip.MinSGIPPduLength {
		return sgip.ErrInvalidPudLength
	}
	b := packet.NewPacketReader(data)
	defer b.Release()
	p.Header = sgip.ReadHeader(b)
	p.Type = sgip.LoginType(b.ReadUint8())
	p.Name = b.ReadCStringN(16)
	p.Password = b.ReadCStringN(16)
	p.Reserved = b.ReadCStringN(8)

	return b.Error()
}

func (p *Bind) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()
	sgip.WriteHeaderNoLength(p.Header, b)
	b.WriteUint8(uint8(p.Type))
	b.WriteFixedLenString(p.Name, 16)
	b.WriteFixedLenString(p.Password, 16)
	b.WriteFixedLenString(p.Reserved, 8)
	return b.BytesWithLength()
}

func (p *Bind) SetSequenceID(id uint32) {
	p.Header.Sequence[2] = id
}

func (b *Bind) GetSequenceID() uint32 {
	return b.Header.Sequence[2]
}

func (b *Bind) GetCommand() sms.ICommander {
	return sgip.SGIP_BIND
}

func (b *Bind) GenEmptyResponse() sms.PDU {
	return &BindResp{
		Header: sgip.NewHeader(0, sgip.SGIP_BIND_REP, b.Sequence[0], b.GetSequenceID()),
	}
}

func (b *Bind) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", b.Header)
	w.Write("Type", b.Type)
	w.Write("Name", b.Name)
	w.Write("Password", b.Password)
	w.Write("Reserved", b.Reserved)

	return w.String()
}

type BindResp struct {
	sgip.Header

	// 1 字节 Bind命令是否成功接收。 0:接收成功 其它:错误码
	Result sgip.RespStatus

	// 8 字节 保留，扩展用
	Reserved string
}

func (p *BindResp) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()
	sgip.WriteHeaderNoLength(p.Header, b)
	b.WriteUint8(uint8(p.Result))
	b.WriteFixedLenString(p.Reserved, 8)
	return b.BytesWithLength()
}

func (p *BindResp) IDecode(data []byte) error {
	b := packet.NewPacketReader(data)
	defer b.Release()
	p.Header = sgip.ReadHeader(b)
	p.Result = sgip.RespStatus(b.ReadUint8())
	p.Reserved = b.ReadCStringN(8)
	return b.Error()
}

func (p *BindResp) SetSequenceID(id uint32) {
	p.Header.Sequence[2] = id
}

func (b *BindResp) GetSequenceID() uint32 {
	return b.Header.Sequence[2]
}

func (b *BindResp) GetCommand() sms.ICommander {
	return sgip.SGIP_BIND_REP
}

func (b *BindResp) GenEmptyResponse() sms.PDU {
	return nil
}

func (b *BindResp) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", b.Header)
	w.Write("Result", b.Result)
	w.Write("Reserved", b.Reserved)

	return w.String()
}
