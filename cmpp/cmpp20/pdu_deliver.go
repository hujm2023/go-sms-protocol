package cmpp20

import (
	"github.com/hujm2023/go-sms-protocol"
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

	p.Header = b.ReadHeader()
	b.ReadNumeric(&p.MsgID)
	p.DestID = b.ReadCStringN(21)
	p.ServiceID = b.ReadCStringN(10)
	b.ReadNumeric(&p.TpPID)
	b.ReadNumeric(&p.TpUDHI)
	b.ReadNumeric(&p.MsgFmt)
	p.SrcTerminalID = b.ReadCStringN(21)
	b.ReadNumeric(&p.RegisteredDeliver)
	b.ReadNumeric(&p.MsgLength)
	p.MsgContent = string(b.ReadNBytes(int(p.MsgLength)))
	p.Reserved = b.ReadCStringN(8)

	return b.Error()
}

func (p *PduDeliver) GetHeader() cmpp.Header {
	return p.Header
}

func (p *PduDeliver) GetSequenceID() uint32 {
	return p.GetHeader().SequenceID
}

func (p *PduDeliver) GetCommandID() cmpp.CommandID {
	return cmpp.CommandDeliver
}

func (p *PduDeliver) GenerateResponseHeader() protocol.PDU {
	resp := &PduDeliverResp{
		Header: cmpp.NewHeader(MaxDeliverRespLength, cmpp.CommandDeliverResp, p.GetSequenceID()),
	}
	return resp
}

func (p *PduDeliver) MaxLength() uint32 {
	return MaxDeliverLength
}

func (p *PduDeliver) SetSequenceID(sid uint32) {
	p.Header.SequenceID = sid
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

	pr.Header = b.ReadHeader()
	b.ReadNumeric(&pr.MsgID)
	b.ReadNumeric(&pr.Result)

	return b.Error()
}

func (pr *PduDeliverResp) GetHeader() cmpp.Header {
	return pr.Header
}

func (pr *PduDeliverResp) GetSequenceID() uint32 {
	return pr.GetHeader().SequenceID
}

func (pr *PduDeliverResp) GetCommandID() cmpp.CommandID {
	return cmpp.CommandDeliverResp
}

func (pr *PduDeliverResp) GenerateResponseHeader() protocol.PDU {
	return nil
}

func (pr *PduDeliverResp) MaxLength() uint32 {
	return MaxDeliverRespLength
}

func (pr *PduDeliverResp) SetSequenceID(sid uint32) {
	pr.Header.SequenceID = sid
}

type SubPduDeliveryContent struct {
	// 8 字节，信息标识SP提交短信（CMPP_SUBMIT）操作时，与SP相连的ISMG产生的Msg_Id。
	MsgID uint64

	// 7 字节. 发送短信的应答结果，含义与SMPP协议要求中stat字段定义相同
	// DELIVRD: Message is delivered to destination
	// EXPIRED: Message validity period has expired
	// DELETED: Message has been deleted.
	// ACCEPTD: Message  is  in  accepted  state(i.e.  has been  manually  readon  behalf  of  the subscriber by customer service)
	// UNKNOWN: Message is in invalid state
	// REJECTD: Message is in a rejected state
	Stat string

	// 10 字节，提交到下游网关时间，YYMMDDHHMM（YY为年的后两位00-99，MM：01-12，DD：01-31，HH：00-23，MM：00-59）
	SubmitTime string

	// 10 字节，收到下游网关状态报告时间，YYMMDDHHMM（YY为年的后两位00-99，MM：01-12，DD：01-31，HH：00-23，MM：00-59）
	DoneTime string

	// 21 字节，目 的 终 端 MSISDN号 码 (SP发 送 CMPP_SUBMIT消息的目标终端)
	DestTerminalID string

	// 4 字节，取自SMSC发送状态报告的消息体中的消息标识。
	SMSCSequence uint32
}

func (s *SubPduDeliveryContent) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter(8 + 7 + 10 + 10 + 21 + 4)
	defer b.Release()

	b.WriteUint64(s.MsgID)
	b.WriteFixedLenString(s.Stat, 7)
	b.WriteFixedLenString(s.SubmitTime, 10)
	b.WriteFixedLenString(s.DoneTime, 10)
	b.WriteFixedLenString(s.DestTerminalID, 21)
	b.WriteUint32(s.SMSCSequence)

	return b.Bytes()
}

func (s *SubPduDeliveryContent) IDecode(data []byte) (err error) {
	r := packet.NewPacketReader(data)
	defer r.Release()

	r.ReadNumeric(&s.MsgID)
	s.Stat = r.ReadCStringN(7)
	s.SubmitTime = r.ReadCStringN(10)
	s.DoneTime = r.ReadCStringN(10)
	s.DestTerminalID = r.ReadCStringN(21)
	r.ReadNumeric(&s.SMSCSequence)

	return r.Error()
}

func (s *SubPduDeliveryContent) SetSequenceID(_ uint32) {
}
