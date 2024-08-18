package cmpp20

import (
	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/packet"
)

type PduDeliver struct {
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

	// 21字节，源终端 MSISDN 号码(状态报告时填为 PduSubmit 消息的目的终端号码)
	SrcTerminalID string `safe_json:"-"`

	// 1 字节，是否为状态报告 --> 0:非状态报告(上行) 1:状态报告(回执)
	RegisteredDeliver uint8

	// 1 字节，消息长度
	MsgLength uint8

	// MsgLength 个字节，消息内容
	MsgContent string

	// 8 字节，保留项
	Reserved string
}

func (p *PduDeliver) IEncode() ([]byte, error) {
	totalLen := HeaderLength + 8 + 21 + 10 + 1 + 1 + 1 + 21 + 1 + 1 + int(p.MsgLength) + 8
	p.TotalLength = uint32(totalLen)
	b := packet.NewPacketWriter(totalLen)
	defer b.Release()

	b.WriteBytes(p.Header.Bytes())
	b.WriteUint64(p.MsgID)
	b.WriteFixedLenString(p.DestID, 21)
	b.WriteFixedLenString(p.ServiceID, 10)
	b.WriteUint8(p.TpPID)
	b.WriteUint8(p.TpUDHI)
	b.WriteUint8(p.MsgFmt)
	b.WriteFixedLenString(p.SrcTerminalID, 21)
	b.WriteUint8(p.RegisteredDeliver)
	b.WriteUint8(p.MsgLength)
	b.WriteString(p.MsgContent)
	b.WriteFixedLenString(p.Reserved, 8)

	return b.Bytes()
}

func (p *PduDeliver) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return ErrInvalidPudLength
	}

	b := packet.NewPacketReader(data)
	defer b.Release()

	p.Header = cmpp.ReadHeader(b)
	p.MsgID = b.ReadUint64()
	p.DestID = b.ReadCStringN(21)
	p.ServiceID = b.ReadCStringN(10)
	p.TpPID = b.ReadUint8()
	p.TpUDHI = b.ReadUint8()
	p.MsgFmt = b.ReadUint8()
	p.SrcTerminalID = b.ReadCStringN(21)
	p.RegisteredDeliver = b.ReadUint8()
	p.MsgLength = b.ReadUint8()
	p.MsgContent = string(b.ReadNBytes(int(p.MsgLength)))
	p.Reserved = b.ReadCStringN(8)

	return b.Error()
}

func (p *PduDeliver) GetSequenceID() uint32 {
	return p.Header.SequenceID
}

func (p *PduDeliver) SetSequenceID(sid uint32) {
	p.Header.SequenceID = sid
}

func (p *PduDeliver) GetCommand() sms.ICommander {
	return cmpp.CommandDeliver
}

func (p *PduDeliver) GenEmptyResponse() sms.PDU {
	return &PduDeliverResp{
		Header: cmpp.Header{
			CommandID:  cmpp.CommandDeliverResp,
			SequenceID: p.GetSequenceID(),
		},
	}
}

func (p *PduDeliver) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", p.Header)
	w.Write("MsgID", p.MsgID)
	w.Write("DestID", p.DestID)
	w.Write("ServiceID", p.ServiceID)
	w.Write("TpPID", p.TpPID)
	w.Write("TpUDHI", p.TpUDHI)
	w.Write("MsgFmt", p.MsgFmt)
	w.Write("SrcTerminalID", p.SrcTerminalID)
	w.Write("RegisteredDeliver", p.RegisteredDeliver)
	w.Write("MsgLength", p.MsgLength)
	w.WriteWithBytes("MsgContent", p.MsgContent)
	w.Write("Reserved", p.Reserved)

	return w.String()
}

// --------------------------------------

type PduDeliverResp struct {
	cmpp.Header

	// 8 字节，PduDeliver 中的 MsgID 字段
	MsgID uint64

	// 1 字节，结果
	// 0:正确
	// 1:消息结构错 2:命令字错 3:消息序号重复 4:消息长度错 5:资费代码错
	// 6:超过最大信息长 7:业务代码错 8: 流量控制错 9~ :其他错误
	Result uint8
}

func (pr *PduDeliverResp) IEncode() ([]byte, error) {
	pr.TotalLength = MaxDeliverRespLength
	b := packet.NewPacketWriter(MaxDeliverRespLength)
	defer b.Release()

	b.WriteBytes(pr.Header.Bytes())
	b.WriteUint64(pr.MsgID)
	b.WriteUint8(pr.Result)

	return b.Bytes()
}

func (pr *PduDeliverResp) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return ErrInvalidPudLength
	}

	b := packet.NewPacketReader(data)
	defer b.Release()

	pr.Header = cmpp.ReadHeader(b)
	pr.MsgID = b.ReadUint64()
	pr.Result = b.ReadUint8()

	return b.Error()
}

func (pr *PduDeliverResp) GetSequenceID() uint32 {
	return pr.Header.SequenceID
}

func (pr *PduDeliverResp) SetSequenceID(sid uint32) {
	pr.Header.SequenceID = sid
}

func (p *PduDeliverResp) GetCommand() sms.ICommander {
	return cmpp.CommandDeliverResp
}

func (p *PduDeliverResp) GenEmptyResponse() sms.PDU {
	return nil
}

func (p *PduDeliverResp) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", p.Header)
	w.Write("MsgID", p.MsgID)
	w.Write("Result", p.Result)

	return w.String()
}
