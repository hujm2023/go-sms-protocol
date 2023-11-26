package cmpp30

import (
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/packet"
)

type Submit struct {
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

	// 32 字节 数字？被计费用户的号码(如本字节填空，则表 示本字段无效，对谁计费参见 Fee_UserType 字段，本字段与 Fee_UserType 字段互斥)
	FeeTerminalID string

	// 1 字节，被计费用户的号码类型，0：真实号码；1：伪码
	FeeTerminalType uint8

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
	ValiDTime string

	// 17 字节，定时发送时间，格式遵循 SMPP3.3 协议
	AtTime string

	// 21 字节，源号码，SP 的服务代码或前缀为服务代码的长号码, 网关将该号码完整的填到 SMPP 协议 Submit_SM 消息相应的 source_addr 字段，
	// 该号码最终在用户手机上显示为短消息 的主叫号码
	// 在我们的实现中，对应 userExt 字段
	SrcID string

	// 1 字节，接收信息的用户数量(小于 100 个用户)
	DestUsrTL uint8

	// 32*DestUsrTL 字节，接收短信的 MSISDN 号码(电话号码)。单个 MSISDN 长度为 32，逐个读取。
	DestTerminalID []string

	// 1字节，接收短信的用户的号码类型，0：真实号码；1：伪码
	DestTerminalType uint8

	// 1 字节，信息长度(MsgFmt=0：<160个字节；其他：<=140个字)
	MsgLength uint8

	// MsgLength 字节，信息内容
	MsgContent string

	// 20字节，点播业务使用的 LinkID，非点播类业务的 MT 流程不使用该字段
	LinkID string
}

func (s *Submit) IDecode(data []byte) error {
	if len(data) < 4 {
		return cmpp.ErrInvalidPudLength
	}
	b := packet.NewPacketReader(data)
	defer b.Release()

	s.Header = cmpp.ReadHeader(b)
	s.MsgID = b.ReadUint64()
	s.PkTotal = b.ReadUint8()
	s.PkNumber = b.ReadUint8()
	s.RegisteredDelivery = b.ReadUint8()
	s.MsgLevel = b.ReadUint8()
	s.ServiceID = b.ReadCStringN(10)
	s.FeeUserType = b.ReadUint8()
	s.FeeTerminalID = b.ReadCStringN(32)
	s.FeeTerminalType = b.ReadUint8()
	s.TpPID = b.ReadUint8()
	s.TpUDHI = b.ReadUint8()
	s.MsgFmt = b.ReadUint8()
	s.MsgSrc = b.ReadCStringN(6)
	s.FeeType = b.ReadCStringN(2)
	s.FeeCode = b.ReadCStringN(6)
	s.ValiDTime = b.ReadCStringN(17)
	s.AtTime = b.ReadCStringN(17)
	s.SrcID = b.ReadCStringN(21)
	s.DestUsrTL = b.ReadUint8()
	s.DestTerminalID = make([]string, s.DestUsrTL)
	for i := 0; i < int(s.DestUsrTL); i++ {
		s.DestTerminalID[i] = b.ReadCStringN(32)
	}
	s.DestTerminalType = b.ReadUint8()
	s.MsgLength = b.ReadUint8()
	s.MsgContent = string(b.ReadNBytes(int(s.MsgLength)))
	s.LinkID = b.ReadCStringN(20)

	return nil
}

func (s *Submit) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()

	cmpp.WriteHeaderNoLength(s.Header, b)
	b.WriteUint64(s.MsgID)
	b.WriteUint8(s.PkTotal)
	b.WriteUint8(s.PkNumber)
	b.WriteUint8(s.RegisteredDelivery)
	b.WriteUint8(s.MsgLevel)
	b.WriteFixedLenString(s.ServiceID, 10)
	b.WriteUint8(s.FeeUserType)
	b.WriteFixedLenString(s.FeeTerminalID, 32)
	b.WriteUint8(s.FeeTerminalType)
	b.WriteUint8(s.TpPID)
	b.WriteUint8(s.TpUDHI)
	b.WriteUint8(s.MsgFmt)
	b.WriteFixedLenString(s.MsgSrc, 6)
	b.WriteFixedLenString(s.FeeType, 2)
	b.WriteFixedLenString(s.FeeCode, 6)
	b.WriteFixedLenString(s.ValiDTime, 17)
	b.WriteFixedLenString(s.AtTime, 17)
	b.WriteFixedLenString(s.SrcID, 21)
	b.WriteUint8(s.DestUsrTL)
	for _, id := range s.DestTerminalID {
		b.WriteFixedLenString(id, 32)
	}
	b.WriteUint8(s.DestTerminalType)
	b.WriteUint8(s.MsgLength)
	b.WriteFixedLenString(s.MsgContent, int(s.MsgLength))
	b.WriteFixedLenString(s.LinkID, 20)

	return b.BytesWithLength()
}

func (s *Submit) SetSequenceID(id uint32) {
	s.Header.SequenceID = id
}

type SubmitResp struct {
	Header cmpp.Header

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
	Result uint32
}

func (s *SubmitResp) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return cmpp.ErrInvalidPudLength
	}

	b := packet.NewPacketReader(data)
	defer b.Release()

	s.Header = cmpp.ReadHeader(b)
	s.MsgID = b.ReadUint64()
	s.Result = b.ReadUint32()

	return b.Error()
}

func (s *SubmitResp) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()

	cmpp.WriteHeaderNoLength(s.Header, b)
	b.WriteUint64(s.MsgID)
	b.WriteUint32(s.Result)

	return b.BytesWithLength()
}

func (s *SubmitResp) SetSequenceID(id uint32) {
	s.Header.SequenceID = id
}
