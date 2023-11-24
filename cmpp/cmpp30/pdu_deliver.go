package cmpp30

import (
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/packet"
)

type Deliver struct {
	cmpp.Header

	/* 8 字节，信息标识，采用 64 位(8 字节)的整数:
	(1)时间(格式为 MMDDHHMMSS，即月日时分秒):bit64~bit39，其中
		bit64~bit61:月份的二进制表示;
		bit60~bit56:日的二进制表示;
		bit55~bit51:小时的二进制表示;
		bit50~bit45:分的二进制表示;
		bit44~bit39:秒的二进制表示;
	(2)短信网关代码:bit38~bit17，把短信 网关的代码转换为整数填写到该字 段中。
	(3)序列号:bit16~bit1，顺序增加，步 长为 1，循环使用。
	各部分如不能填满，左补零，右对齐。
	*/
	MsgID uint64

	// 21字节，目标号码
	// SP 的服务代码，一般 4--6 位，或者是前缀为服务代码的长号码;该号码是手机用户短消息的被叫号码。
	DestID string

	// 10 字节，业务类型，是数字、字母和符号的 组合。
	ServiceID string

	// 1 字节，GSM 协议类型。详情请参考 GSM3.40 中的9.2.3.9: https://www.etsi.org/deliver/etsi_gts/03/0340/05.03.00_60/gsmts_0340v050300p.pdf
	TpPID uint8

	// 1 字节，GSM 协议类型。详细是解释请参考 GSM03.40 中的 9.2.3.23,仅使用 1 位，右对齐
	TpUDHI uint8

	// 1 字节，信息格式 --> 0:ASCII 串 3:短信写卡操作 4:二进制信息 8:UCS2 编码 15:含GB汉字
	MsgFmt uint8

	// 32字节，源终端 MSISDN 号码(状态报告时填为 PduSubmit 消息的目的终端号码)
	SrcTerminalID string `safe_json:"-"`

	// 1 字节，源终端号码类型，0：真实号码；1：伪码
	SrcTerminalType uint8

	// 1 字节，是否为状态报告 --> 0:非状态报告(上行) 1:状态报告(回执)
	RegisteredDeliver uint8

	// 1 字节，消息长度
	MsgLength uint8

	// MsgLength 个字节，消息内容
	MsgContent string

	// 20 字节，点播业务使用的 LinkID，非点播类业务的 MT 流程不使用该字段
	LinkID string
}

func (d *Deliver) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return cmpp.ErrInvalidPudLength
	}

	b := packet.NewPacketReader(data)
	defer b.Release()

	d.Header = cmpp.ReadHeader(b)
	d.MsgID = b.ReadUint64()
	d.DestID = b.ReadCStringN(21)
	d.ServiceID = b.ReadCStringN(10)
	d.TpPID = b.ReadUint8()
	d.TpUDHI = b.ReadUint8()
	d.MsgFmt = b.ReadUint8()
	d.SrcTerminalID = b.ReadCStringN(32)
	d.SrcTerminalType = b.ReadUint8()
	d.RegisteredDeliver = b.ReadUint8()
	d.MsgLength = b.ReadUint8()
	d.MsgContent = string(b.ReadNBytes(int(d.MsgLength)))
	d.LinkID = b.ReadCStringN(20)

	return b.Error()
}

func (d *Deliver) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()

	b.WriteUint32(uint32(d.Header.CommandID))
	b.WriteUint32(d.Header.SequenceID)
	b.WriteUint64(d.MsgID)
	b.WriteFixedLenString(d.DestID, 21)
	b.WriteFixedLenString(d.ServiceID, 10)
	b.WriteUint8(d.TpPID)
	b.WriteUint8(d.TpUDHI)
	b.WriteUint8(d.MsgFmt)
	b.WriteFixedLenString(d.SrcTerminalID, 32)
	b.WriteUint8(d.SrcTerminalType)
	b.WriteUint8(d.RegisteredDeliver)
	b.WriteUint8(d.MsgLength)
	b.WriteFixedLenString(d.MsgContent, int(d.MsgLength))
	b.WriteFixedLenString(d.LinkID, 20)

	return b.BytesWithLength()
}

func (d *Deliver) SetSequenceID(id uint32) {
	d.Header.SequenceID = id
}

type DeliverResp struct {
	Header cmpp.Header
	// 8 字节，信息标识（CMPP_DELIVER 中的 Msg_Id 字段）
	MsgID uint64

	// 4 字节, 回执结果
	// 0:正确
	// 1:消息结构错 2:命令字错 3:消息序号重复 4:消息长度错 5:资费代码错
	// 6:超过最大信息长 7:业务代码错 8: 流量控制错 9~ :其他错误
	Result uint32
}

func (d *DeliverResp) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return cmpp.ErrInvalidPudLength
	}
	b := packet.NewPacketReader(data)
	defer b.Release()

	d.Header = cmpp.ReadHeader(b)
	d.MsgID = b.ReadUint64()
	d.Result = b.ReadUint32()

	return b.Error()
}

func (d *DeliverResp) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()

	b.WriteUint32(uint32(d.Header.CommandID))
	b.WriteUint32(d.Header.SequenceID)
	b.WriteUint64(d.MsgID)
	b.WriteUint32(d.Result)

	return b.BytesWithLength()
}

func (d *DeliverResp) SetSequenceID(id uint32) {
	d.Header.SequenceID = id
}
