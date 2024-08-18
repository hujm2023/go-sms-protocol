package sgip12

import (
	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/packet"
	"github.com/hujm2023/go-sms-protocol/sgip"
)

type Submit struct {
	sgip.Header

	// body 在 SP 和 SMG 的通信中，SP 用 Submit 命令向 SMG 提交 MT 短消息，SMG 返回响应。

	// 21 字节 SP 的接入号码
	SpNumber string

	// 21 字节 付费号码，字符，手机号码前加“86”国别标志;当且仅当群发 且对用户收费时为空;如果为空，则该条短消息产生的费用由 UserNumber 代表的用户支付;
	// 如果为全零字符串 “000000000000000000000”，表示该条短消息产生的费用由 SP 支付。
	ChargeNumber string

	// 1 字节 接收短消息的手机数量，取值范围 1 至 100
	UserCount uint8

	//	21*UserCount  字节 一个或多个接收该短消息的手机号，手机号之间用逗号(,)隔开， 字 符 ， 手 机 号 码 前 加 “ 86 ” 国 别 标 志 ，
	//	如 8613001125453,8613001132345
	UserNumber []string // 接收该短消息的手机号，该字段重复UserCount指定的次数，手机号码前加“86”国别标志

	// 5 字节 企业代码，取值范围 0-99999，字符
	CorpID string

	// 10 字节 业务代码，由 SP 定义
	ServiceType string

	// 1 字节 计费类型
	FeeType uint8

	// 6 字节 取值范围 0-99999，该条短消息的收费值，单位为 分，由 SP 定义 对于包月制收费的用户，该值为月租费的值
	FeeValue string

	// 6 字节 取值范围 0-99999，赠送用户的话费，单位为分， 由 SP 定义，特指由 SP 向用户发送广告时的赠送话费
	GivenValue string

	// 1 字节 代收费标志，0:应收;1:实收
	AgentFlag uint8

	/*
		1 字节
		引起 MT 消息的原因
		0-MO 点播引起的第一条 MT 消息;
		1-MO 点播引起的非第一条 MT 消息;
		2-非 MO 点播引起的 MT 消息;
		3-系统反馈引起的 MT 消息。
	*/
	MorelatetoMTFlag uint8

	// 1 字节 优先级 0-9 从低到高，默认为 0
	Priority uint8

	/*
		16 字节
		短消息寿命的终止时间，如果为空，表示使用短消 息中心的缺省值。时间内容为 16 个字符，格式
		为”yymmddhhmmsstnnp” ，其中“tnnp”取 固定值“032+”，即默认系统为北京时间
	*/
	ExpireTime string

	/*
		16 字节
		短消息定时发送的时间，如果为空，表示立刻发送 该短消息。时间内容为 16 个字符，格式为
		“yymmddhhmmsstnnp” ，其中“tnnp”取 固定值“032+”，即默认系统为北京时间
	*/
	ScheduleTime string

	/*
		1 字节
		状态报告标记
		0-该条消息只有最后出错时要返回状态报告
		1-该条消息无论最后是否成功都要返回状态报告
		2-该条消息不需要返回状态报告
		3-该条消息仅携带包月计费信息，不下发给用户， 要返回状态报告
		其它-保留
		缺省设置为 0
	*/
	ReportFlag uint8

	TpPid  uint8
	TpUdhi uint8

	/*
		1 字节
		短消息的编码格式。
		0:纯 ASCII 字符串
		3:写卡操作 4:二进制编码
		8:UCS2 编码
		15: GBK 编码
		其它
	*/
	MessageCoding uint8

	// 1 字节 信息类型: 0-短消息信息 其它:待定
	MessageType uint8

	// 4 字节 短消息的长度
	MessageLength uint32

	// MessageLength 字节 短消息的内容
	MessageContent string

	// 8 字节 保留，扩展用
	Reserved string
}

// 实现真正的PDU接口

func (p *Submit) IDecode(data []byte) error {
	if len(data) < sgip.MinSGIPPduLength {
		return sgip.ErrInvalidPudLength
	}
	b := packet.NewPacketReader(data)
	defer b.Release()

	p.Header = sgip.ReadHeader(b)
	p.SpNumber = b.ReadCStringN(21)
	p.ChargeNumber = b.ReadCStringN(21)
	p.UserCount = b.ReadUint8()
	for i := 0; i < int(p.UserCount); i++ {
		nubmer := b.ReadCStringN(21)
		p.UserNumber = append(p.UserNumber, nubmer)
	}
	p.CorpID = b.ReadCStringN(5)
	p.ServiceType = b.ReadCStringN(10)
	p.FeeType = b.ReadUint8()
	p.FeeValue = b.ReadCStringN(6)
	p.GivenValue = b.ReadCStringN(6)
	p.AgentFlag = b.ReadUint8()
	p.MorelatetoMTFlag = b.ReadUint8()
	p.Priority = b.ReadUint8()
	p.ExpireTime = b.ReadCStringN(16)
	p.ScheduleTime = b.ReadCStringN(16)
	p.ReportFlag = b.ReadUint8()
	p.TpPid = b.ReadUint8()
	p.TpUdhi = b.ReadUint8()
	p.MessageCoding = b.ReadUint8()
	p.MessageType = b.ReadUint8()
	p.MessageLength = b.ReadUint32()
	p.MessageContent = string(b.ReadNBytes(int(p.MessageLength)))
	p.Reserved = b.ReadCStringN(8)
	return b.Error()
}

func (p *Submit) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()
	sgip.WriteHeaderNoLength(p.Header, b)
	b.WriteFixedLenString(p.SpNumber, 21)
	b.WriteFixedLenString(p.ChargeNumber, 21)
	// 纠正一下可能才出现的 userCount 和 len(UserNumber)可能不一致的小问题
	if len(p.UserNumber) != int(p.UserCount) {
		p.UserCount = uint8(len(p.UserNumber))
	}
	b.WriteUint8(p.UserCount)
	// serialize UserNumber
	for i := 0; i < int(p.UserCount); i++ {
		b.WriteFixedLenString(p.UserNumber[i], 21)
	}
	b.WriteFixedLenString(p.CorpID, 5)
	b.WriteFixedLenString(p.ServiceType, 10)
	b.WriteUint8(p.FeeType)
	b.WriteFixedLenString(p.FeeValue, 6)
	b.WriteFixedLenString(p.GivenValue, 6)
	b.WriteUint8(p.AgentFlag)
	b.WriteUint8(p.MorelatetoMTFlag)
	b.WriteUint8(p.Priority)
	b.WriteFixedLenString(p.ExpireTime, 16)
	b.WriteFixedLenString(p.ScheduleTime, 16)
	b.WriteUint8(p.ReportFlag)
	b.WriteUint8(p.TpPid)
	b.WriteUint8(p.TpUdhi)
	b.WriteUint8(p.MessageCoding)
	b.WriteUint8(p.MessageType)
	b.WriteUint32(p.MessageLength)
	b.WriteFixedLenString(p.MessageContent, int(p.MessageLength))
	b.WriteFixedLenString(p.Reserved, 8)

	return b.BytesWithLength()
}

func (p *Submit) SetSequenceID(id uint32) {
	p.Header.Sequence[2] = id
}

func (s *Submit) GetSequenceID() uint32 {
	return s.Header.Sequence[2]
}

func (s *Submit) GetCommand() sms.ICommander {
	return sgip.SGIP_SUBMIT
}

func (s *Submit) GenEmptyResponse() sms.PDU {
	return &SubmitResp{
		Header: sgip.NewHeader(sgip.MaxHeaderRespLength, sgip.SGIP_SUBMIT_REP, s.GetSequenceID(), s.GetSequenceID()),
	}
}

func (s *Submit) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", s.Header)
	w.Write("SpNumber", s.SpNumber)
	w.Write("ChargeNumber", s.ChargeNumber)
	w.Write("UserCount", s.UserCount)
	w.Write("UserNumber", s.UserNumber)
	w.Write("CorpID", s.CorpID)
	w.Write("ServiceType", s.ServiceType)
	w.Write("FeeType", s.FeeType)
	w.Write("FeeValue", s.FeeValue)
	w.Write("GivenValue", s.GivenValue)
	w.Write("AgentFlag", s.AgentFlag)
	w.Write("MorelatetoMTFlag", s.MorelatetoMTFlag)
	w.Write("Priority", s.Priority)
	w.Write("ExpireTime", s.ExpireTime)
	w.Write("ScheduleTime", s.ScheduleTime)
	w.Write("ReportFlag", s.ReportFlag)
	w.Write("TpPid", s.TpPid)
	w.Write("TpUdhi", s.TpUdhi)
	w.Write("MessageCoding", s.MessageCoding)
	w.Write("MessageType", s.MessageType)
	w.Write("MessageLength", s.MessageLength)
	w.WriteWithBytes("MessageContent", s.MessageContent)
	w.Write("Reserved", s.Reserved)

	return w.String()
}

type SubmitResp struct {
	sgip.Header
	// 1 字节 Submit命令是否成功接收。 0:接收成功 其它:错误码
	Result sgip.RespStatus
	// 8 字节 保留，扩展用
	Reserved string
}

func (p *SubmitResp) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()
	sgip.WriteHeaderNoLength(p.Header, b)
	b.WriteUint8(uint8(p.Result))
	b.WriteFixedLenString(p.Reserved, 8)
	return b.BytesWithLength()
}

func (p *SubmitResp) IDecode(data []byte) error {
	b := packet.NewPacketReader(data)
	defer b.Release()
	p.Header = sgip.ReadHeader(b)
	p.Result = sgip.RespStatus(b.ReadUint8())
	p.Reserved = b.ReadCStringN(8)
	return b.Error()
}

func (p *SubmitResp) SetSequenceID(id uint32) {
	p.Header.Sequence[2] = id
}

func (s *SubmitResp) GetSequenceID() uint32 {
	return s.Header.Sequence[2]
}

func (s *SubmitResp) GetCommand() sms.ICommander {
	return sgip.SGIP_SUBMIT_REP
}

func (s *SubmitResp) GenEmptyResponse() sms.PDU {
	return nil
}

func (s *SubmitResp) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", s.Header)
	w.Write("Result", s.Result)
	w.Write("Reserved", s.Reserved)

	return w.String()
}
