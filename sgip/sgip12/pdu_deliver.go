package sgip12

import (
	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/packet"
	"github.com/hujm2023/go-sms-protocol/sgip"
)

type Deliver struct {
	sgip.Header

	// body
	// 21 字节 发送短消息的用户手机号，手机号码前加“86”国 别标志
	UserNumber string

	// 21 字节 SP 的接入号码
	SPNumber string

	// 1 字节 GSM 协议类型。详细解释请参考 GSM03.40 中的 9.2.3.9
	TpPid uint8

	// 1 字节 GSM 协议类型。详细解释请参考 GSM03.40 中的 9.2.3.23,仅使用 1 位，右对齐
	TpUdhi uint8

	// 1 字节
	// 短消息的编码格式。 0:纯 ASCII 字符串 3:写卡操作 4:二进制编码 8:UCS2 编码 15: GBK 编码
	// 其它参见 GSM3.38 第 4 节:SMS Data Coding Scheme
	MessageCoding uint8

	// 4 字节 短消息的长度
	MessageLength uint32

	// MessageLength 字节 短消息的内容
	MessageContent []byte

	// 8 字节 保留，扩展用
	Reserved string
}

func (p *Deliver) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	sgip.WriteHeaderNoLength(p.Header, b)
	defer b.Release()
	b.WriteFixedLenString(p.UserNumber, 21)
	b.WriteFixedLenString(p.SPNumber, 21)
	b.WriteUint8(p.TpPid)
	b.WriteUint8(p.TpUdhi)
	b.WriteUint8(p.MessageCoding)
	b.WriteUint32(p.MessageLength)
	b.WriteBytes(p.MessageContent)
	b.WriteFixedLenString(p.Reserved, 8)

	return b.BytesWithLength()
}

func (p *Deliver) IDecode(data []byte) error {
	if len(data) < sgip.MinSGIPPduLength {
		return sgip.ErrInvalidPudLength
	}
	b := packet.NewPacketReader(data)
	defer b.Release()
	p.Header = sgip.ReadHeader(b)
	p.UserNumber = b.ReadCStringN(21)
	p.SPNumber = b.ReadCStringN(21)
	p.TpPid = b.ReadUint8()
	p.TpUdhi = b.ReadUint8()
	p.MessageCoding = b.ReadUint8()
	p.MessageLength = b.ReadUint32()
	p.MessageContent = b.ReadNBytes(int(p.MessageLength))
	p.Reserved = b.ReadCStringN(8)
	return b.Error()
}

func (p *Deliver) SetSequenceID(id uint32) {
	p.Header.Sequence[2] = id
}

func (d *Deliver) GetSequenceID() uint32 {
	return d.Header.Sequence[2]
}

func (d *Deliver) GetCommand() sms.ICommander {
	return sgip.SGIP_DELIVER
}

func (d *Deliver) GenEmptyResponse() sms.PDU {
	return &DeliverResp{
		Header: sgip.NewHeader(0, sgip.SGIP_DELIVER_REP, d.Header.Sequence[0], d.GetSequenceID()),
	}
}

func (d *Deliver) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", d.Header)
	w.Write("UserNumber", d.UserNumber)
	w.Write("SPNumber", d.SPNumber)
	w.Write("TpPid", d.TpPid)
	w.Write("TpUdhi", d.TpUdhi)
	w.Write("MessageCoding", d.MessageCoding)
	w.Write("MessageLength", d.MessageLength)
	w.WriteWithBytes("MessageContent", d.MessageContent)
	w.Write("Reserved", d.Reserved)

	return w.String()
}

type DeliverResp struct {
	sgip.Header

	// 1 字节 Deliver命令是否成功接收。 0:接收成功 其它:错误码
	Result sgip.RespStatus

	// 8 字节 保留，扩展用
	Reserved string
}

func (p *DeliverResp) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()
	sgip.WriteHeaderNoLength(p.Header, b)
	b.WriteUint8(uint8(p.Result))
	b.WriteFixedLenString(p.Reserved, 8)
	return b.BytesWithLength()
}

func (p *DeliverResp) IDecode(data []byte) error {
	b := packet.NewPacketReader(data)
	defer b.Release()
	p.Header = sgip.ReadHeader(b)
	p.Result = sgip.RespStatus(b.ReadUint8())
	p.Reserved = b.ReadCStringN(8)
	return b.Error()
}

func (p *DeliverResp) SetSequenceID(id uint32) {
	p.Header.Sequence[2] = id
}

func (d *DeliverResp) GetSequenceID() uint32 {
	return d.Header.Sequence[2]
}

func (d *DeliverResp) GetCommand() sms.ICommander {
	return sgip.SGIP_DELIVER_REP
}

func (d *DeliverResp) GenEmptyResponse() sms.PDU {
	return nil
}

func (d *DeliverResp) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", d.Header)
	w.Write("Result", d.Result)
	w.Write("Reserved", d.Reserved)

	return w.String()
}
