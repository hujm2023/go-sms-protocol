package cmpp20

import (
	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/packet"
)

type PduSubmit struct {
	cmpp.Header

	// 8 字节，信息标识，由 SP 侧短信网关本身产生， 本处填空。
	MsgID uint64

	// 1 字节，相同 MsgID 的条数，从 1 开始
	PkTotal uint8

	// 1 字节，相同 MsgID 的序号，从 1 开始
	PkNumber uint8

	// 1 字节，是否要求返回状态确认报告: 0不需要 1需要 2产生SMC话单(该类型短信仅供网关计费使用，不发 送给目的终端)
	RegisteredDelivery uint8

	// 1 字节，信息级别
	MsgLevel uint8

	// 10 字节，业务类型，是数字、字母和符号的组合。
	ServiceID string

	// 1 字节，计费用户类型字段：0:对目的终端MSISDN计费; 1:对源终端MSISDN计费; 2:对SP计费;3:表示本字段无效，对谁计费参见 Fee_terminal_Id 字段。
	FeeUserType uint8

	// 21 字节 数字？被计费用户的号码(如本字节填空，则表 示本字段无效，对谁计费参见 Fee_UserType 字段，本字段与 Fee_UserType 字段互斥)
	FeeTerminalID string

	// 1 字节，GSM 协议类型。详情请参考 GSM3.40 中的9.2.3.9: https://www.etsi.org/deliver/etsi_gts/03/0340/05.03.00_60/gsmts_0340v050300p.pdf
	TpPID uint8

	// 1 字节，GSM 协议类型。详细是解释请参考 GSM03.40 中的 9.2.3.23,仅使用 1 位，右对齐
	TpUDHI uint8

	// 1 字节，信息格式 0:ASCII 串 3:短信写卡操作 4:二进制信息 8:UCS2 编码 15:含GB汉字
	MsgFmt uint8

	// 6 字节，信息内容来源(SpID)
	MsgSrc string

	// 2 字节，资费类别。
	// 01:对“计费用户号码”免费 02:对“计费用户号码”按条计信息费 03:对“计费用户号码”按包月收取信息费
	// 04:对“计费用户号码”的信息费封顶 05:对“计费用户号码”的收费是由 SP 实现
	FeeType string

	// 6 字节，资费代码(以分为单位)
	FeeCode string

	// 17 字节，存活有效期，格式遵循 SMPP3.3 协议
	ValIDTime string

	// 17 字节，定时发送时间，格式遵循 SMPP3.3 协议
	AtTime string

	// 21 字节，源号码，SP 的服务代码或前缀为服务代码的长号码, 网关将该号码完整的填到 SMPP 协议 Submit_SM 消息相应的 source_addr 字段，
	// 该号码最终在用户手机上显示为短消息 的主叫号码
	// 在我们的实现中，对应 userExt 字段
	SrcID string

	// 1 字节，接收信息的用户数量(小于 100 个用户)
	DestUsrTL uint8

	// 21*DestUsrTL 字节，接收短信的 MSISDN 号码(电话号码)。单个 MSISDN 长度为 21，逐个读取。
	DestTerminalID []string

	// 1 字节，信息长度(MsgFmt=0：<160个字节；其他：<=140个字)
	MsgLength uint8

	// MsgLength 字节，信息内容
	MsgContent string

	// 8 字节，保留
	Reserve string
}

func (p *PduSubmit) IEncode() ([]byte, error) {
	if p.PkTotal == 0 && p.PkNumber == 0 {
		p.PkTotal, p.PkNumber = 1, 1
	}
	p.TotalLength = HeaderLength + // header
		116 + // DestUsrTL 之前的字段长度
		1 + uint32(21*p.DestUsrTL) + // DestUsrTL 和 DestTerminalID
		1 + uint32(p.MsgLength) + // MsgLength 和 MsgContent
		8 // reversed

	b := packet.NewPacketWriter(int(p.TotalLength))
	defer b.Release()

	b.WriteBytes(p.Header.Bytes())
	b.WriteUint64(p.MsgID)
	b.WriteUint8(p.PkTotal)
	b.WriteUint8(p.PkNumber)
	b.WriteUint8(p.RegisteredDelivery)
	b.WriteUint8(p.MsgLevel)
	b.WriteFixedLenString(p.ServiceID, 10)
	b.WriteUint8(p.FeeUserType)
	b.WriteFixedLenString(p.FeeTerminalID, 21)
	b.WriteUint8(p.TpPID)
	b.WriteUint8(p.TpUDHI)
	b.WriteUint8(p.MsgFmt)
	b.WriteFixedLenString(p.MsgSrc, 6)
	b.WriteFixedLenString(p.FeeType, 2)
	b.WriteFixedLenString(p.FeeCode, 6)
	b.WriteFixedLenString(p.ValIDTime, 17)
	b.WriteFixedLenString(p.AtTime, 17)
	b.WriteFixedLenString(p.SrcID, 21)
	b.WriteUint8(p.DestUsrTL)
	for _, dest := range p.DestTerminalID {
		b.WriteFixedLenString(dest, 21)
	}
	b.WriteUint8(p.MsgLength)
	b.WriteString(p.MsgContent)
	b.WriteFixedLenString(p.Reserve, 8)

	return b.Bytes()
}

func (p *PduSubmit) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return ErrInvalidPudLength
	}
	b := packet.NewPacketReader(data)
	defer b.Release()

	p.Header = cmpp.ReadHeader(b)
	p.MsgID = b.ReadUint64()
	p.PkTotal = b.ReadUint8()
	p.PkNumber = b.ReadUint8()
	p.RegisteredDelivery = b.ReadUint8()
	p.MsgLevel = b.ReadUint8()
	p.ServiceID = b.ReadCStringN(10)
	p.FeeUserType = b.ReadUint8()
	p.FeeTerminalID = b.ReadCStringN(21)
	p.TpPID = b.ReadUint8()
	p.TpUDHI = b.ReadUint8()
	p.MsgFmt = b.ReadUint8()
	p.MsgSrc = b.ReadCStringN(6)
	p.FeeType = b.ReadCStringN(2)
	p.FeeCode = b.ReadCStringN(6)
	p.ValIDTime = b.ReadCStringN(17)
	p.AtTime = b.ReadCStringN(17)
	p.SrcID = b.ReadCStringN(21)
	p.DestUsrTL = b.ReadUint8()
	p.DestTerminalID = make([]string, p.DestUsrTL)
	for i := 0; i < int(p.DestUsrTL); i++ {
		p.DestTerminalID[i] = b.ReadCStringN(21)
	}
	p.MsgLength = b.ReadUint8()
	p.MsgContent = string(b.ReadNBytes(int(p.MsgLength)))
	p.Reserve = b.ReadCStringN(8)

	return b.Error()
}

func (p *PduSubmit) GetSequenceID() uint32 {
	return p.Header.SequenceID
}

func (p *PduSubmit) SetSequenceID(id uint32) {
	p.Header.SequenceID = id
}

func (p *PduSubmit) MaxLength() uint32 {
	return MaxSubmitLength
}

func (p *PduSubmit) GetCommand() sms.ICommander {
	return cmpp.CommandSubmit
}

func (p *PduSubmit) GenEmptyResponse() sms.PDU {
	return &PduSubmitResp{
		Header: cmpp.Header{
			CommandID:  cmpp.CommandSubmitResp,
			SequenceID: p.GetSequenceID(),
		},
	}
}

func (p *PduSubmit) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", p.Header)
	w.Write("MsgID", p.MsgID)
	w.Write("PkTotal", p.PkTotal)
	w.Write("PkNumber", p.PkNumber)
	w.Write("RegisteredDelivery", p.RegisteredDelivery)
	w.Write("MsgLevel", p.MsgLevel)
	w.Write("ServiceID", p.ServiceID)
	w.Write("FeeUserType", p.FeeUserType)
	w.Write("FeeTerminalID", p.FeeTerminalID)
	w.Write("TpPID", p.TpPID)
	w.Write("TpUDHI", p.TpUDHI)
	w.Write("MsgFmt", p.MsgFmt)
	w.Write("MsgSrc", p.MsgSrc)
	w.Write("FeeType", p.FeeType)
	w.Write("FeeCode", p.FeeCode)
	w.Write("ValIDTime", p.ValIDTime)
	w.Write("AtTime", p.AtTime)
	w.Write("SrcID", p.SrcID)
	w.Write("DestUsrTL", p.DestUsrTL)
	w.Write("DestTerminalID", p.DestTerminalID)
	w.Write("MsgLength", p.MsgLength)
	w.WriteWithBytes("MsgContent", p.MsgContent)
	w.Write("Reserve", p.Reserve)

	return w.String()
}

// --------------

type PduSubmitResp struct {
	cmpp.Header

	// 8 字节，信息标识，生成算法如下:
	// 采用 64 位(8 字节)的整数:
	// 		(1)时间(格式为 MMDDHHMMSS，即 月日时分秒)
	// 			:bit64~bit39，其中
	// 				bit64~bit61:月份的二进制表示;
	// 				bit60~bit56:日的二进制表示;
	// 				bit55~bit51:小时的二进制表示;
	// 				bit50~bit45:分的二进制表示;
	// 				bit44~bit39:秒的二进制表示;
	// 		(2)短信网关代码:bit38~bit17，把短信网关的代码转换为整数填写到该字段中。
	// 		(3)序列号:bit16~bit1，顺序增加，步 长为 1，循环使用。
	//
	// 各部分如不能填满，左补零，右对齐。
	// (SP 根据请求和应答消息的 Sequence_Id 一致性就可得到 CMPP_Submit 消息的 Msg_Id)
	MsgID uint64

	// 1 字节，提交结果
	// 0:正确 1:消息结构错 2:命令字错 3:消息序号重复 4:消息长度错 5:资费代码错
	// 6:超过最大信息长 7:业务代码错 8:流量控制错 9~:其他错误
	Result uint8
}

func (pr *PduSubmitResp) IEncode() ([]byte, error) {
	pr.TotalLength = MaxSubmitRespLength
	b := packet.NewPacketWriter(MaxSubmitRespLength)
	defer b.Release()

	b.WriteBytes(pr.Header.Bytes())
	b.WriteUint64(pr.MsgID)
	b.WriteUint8(pr.Result)

	return b.Bytes()
}

func (pr *PduSubmitResp) IDecode(data []byte) error {
	b := packet.NewPacketReader(data)
	defer b.Release()

	pr.Header = cmpp.ReadHeader(b)
	pr.MsgID = b.ReadUint64()
	pr.Result = b.ReadUint8()

	return b.Error()
}

func (pr *PduSubmitResp) GetSequenceID() uint32 {
	return pr.Header.SequenceID
}

func (pr *PduSubmitResp) SetSequenceID(id uint32) {
	pr.Header.SequenceID = id
}

func (p *PduSubmitResp) GetCommand() sms.ICommander {
	return cmpp.CommandSubmitResp
}

func (p *PduSubmitResp) GenEmptyResponse() sms.PDU {
	return nil
}

func (p *PduSubmitResp) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", p.Header)
	w.Write("MsgID", p.MsgID)
	w.Write("Result", p.Result)

	return w.String()
}

var submitResultString = map[uint8]string{
	0: "正确",
	1: "消息结构错",
	2: "命令字错",
	3: "消息序号重复",
	4: "消息长度错",
	5: "资费代码错",
	6: "超过最大信息长",
	7: "业务代码错",
	8: "流量控制错",
	9: "其他错误",
}

// SubmitRespResultString ...
func SubmitRespResultString(n uint8) string {
	v, ok := submitResultString[n]
	if ok {
		return v
	}
	return "其他错误"
}
