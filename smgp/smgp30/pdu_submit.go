package smgp30

import (
	"encoding/hex"

	"github.com/samber/lo"

	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/packet"
	"github.com/hujm2023/go-sms-protocol/smgp"
)

type Submit struct {
	smgp.Header

	// 短消息类型
	MsgType uint8

	// sp是否要求返回状态报告
	NeedReport uint8

	// 短消息发送优先级
	Priority uint8

	// 业务代码 10字节
	ServiceID string

	// 收费类型 2
	// 对于 MO 消息或点对点短消息，该字段无效。对于 MT 消息，该字段用法如下：
	// 00＝免费，此时 FixedFee 和 FeeCode 无效；
	// 01＝按条计信息费，此时 FeeCode 表示每条费用，FixedFee 无效；
	// 02＝按包月收取信息费，此时 FeeCode 无效，FixedFee 表示包月费用
	// 03＝按封顶收取信息费，
	FeeType string

	// 资费代码 6
	// 每条短消息费率，单位为“分”

	FeeCode string

	// 包月费/封顶费 6
	FixedFee string

	// 短消息格式
	MsgFormat uint8

	// 短消息有效时间 17
	ValidTime string

	// 短消息定时发送时间 17
	AtTime string

	// 短消息发送方号码 21
	// SrcTermID 格式为“118＋SP 服务代码＋其它（可选）”，例如 SP 服务代码为 1234 时，SrcTermID 可以为 1181234

	SrcTermID string

	// 计费用户号码 21
	ChargeTermID string

	// 短消息接收号码总数
	DestTermIDCount uint8

	// 短消息接收号码 21*count
	DestTermID []string

	// 短消息长度
	MsgLength uint8

	// 短消息内容
	MsgContent []byte

	// 保留
	Reserve string

	// 可选字段
	Options smgp.Options
}

func (s *Submit) IDecode(data []byte) error {
	if len(data) < 4 {
		return smgp.ErrInvalidPudLength
	}
	b := packet.NewPacketReader(data)
	defer b.Release()

	s.Header = smgp.ReadHeader(b)
	s.MsgType = b.ReadUint8()
	s.NeedReport = b.ReadUint8()
	s.Priority = b.ReadUint8()
	s.ServiceID = b.ReadCStringN(10)
	s.FeeType = b.ReadCStringN(2)
	s.FeeCode = b.ReadCStringN(6)
	s.FixedFee = b.ReadCStringN(6)
	s.MsgFormat = b.ReadUint8()
	s.ValidTime = b.ReadCStringN(17)
	s.AtTime = b.ReadCStringN(17)
	s.SrcTermID = b.ReadCStringN(21)
	s.ChargeTermID = b.ReadCStringN(21)
	s.DestTermIDCount = b.ReadUint8()
	for i := 0; i < int(s.DestTermIDCount); i++ {
		tmp := b.ReadCStringN(21)
		s.DestTermID = append(s.DestTermID, tmp)
	}
	s.MsgLength = b.ReadUint8()
	s.MsgContent = b.ReadNBytes(int(s.MsgLength))
	s.Reserve = b.ReadCStringN(8)

	var parseErr error
	s.Options, parseErr = smgp.ParseOptions(b.Bytes())
	decodeErr := lo.Ternary(b.Error() != nil, b.Error(), parseErr)
	return decodeErr
}

func (s *Submit) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()

	smgp.WriteHeaderNoLength(s.Header, b)
	b.WriteUint8(s.MsgType)
	b.WriteUint8(s.NeedReport)
	b.WriteUint8(s.Priority)
	b.WriteFixedLenString(s.ServiceID, 10)
	b.WriteFixedLenString(s.FeeType, 2)
	b.WriteFixedLenString(s.FeeCode, 6)
	b.WriteFixedLenString(s.FixedFee, 6)
	b.WriteUint8(s.MsgFormat)
	b.WriteFixedLenString(s.ValidTime, 17)
	b.WriteFixedLenString(s.AtTime, 17)
	b.WriteFixedLenString(s.SrcTermID, 21)
	b.WriteFixedLenString(s.ChargeTermID, 21)
	b.WriteUint8(s.DestTermIDCount)
	for _, id := range s.DestTermID {
		b.WriteFixedLenString(id, 21)
	}
	b.WriteUint8(s.MsgLength)
	b.WriteBytes(s.MsgContent)
	b.WriteFixedLenString(s.Reserve, 8)
	b.WriteBytes(s.Options.Serialize())

	return b.BytesWithLength()
}

func (s *Submit) SetSequenceID(id uint32) {
	s.Header.SequenceID = id
}

func (s *Submit) GetSequenceID() uint32 {
	return s.Header.SequenceID
}

func (s *Submit) GetCommand() sms.ICommander {
	return smgp.CommandSubmit
}

func (s *Submit) GenEmptyResponse() sms.PDU {
	return &SubmitResp{
		Header: smgp.NewHeader(0, smgp.CommandSubmitResp, s.GetSequenceID()),
	}
}

func (s *Submit) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", s.Header)
	w.Write("MsgType", s.MsgType)
	w.Write("NeedReport", s.NeedReport)
	w.Write("Priority", s.Priority)
	w.Write("ServiceID", s.ServiceID)
	w.Write("FeeType", s.FeeType)
	w.Write("FeeCode", s.FeeCode)
	w.Write("FixedFee", s.FixedFee)
	w.Write("MsgFormat", s.MsgFormat)
	w.Write("ValidTime", s.ValidTime)
	w.Write("AtTime", s.AtTime)
	w.Write("SrcTermID", s.SrcTermID)
	w.Write("ChargeTermID", s.ChargeTermID)
	w.Write("DestTermIDCount", s.DestTermIDCount)
	w.Write("DestTermID", s.DestTermID)
	w.Write("MsgLength", s.MsgLength)
	w.WriteWithBytes("MsgContent", s.MsgContent)
	w.Write("Reserve", s.Reserve)
	w.OmitWrite("Options", s.Options.String())

	return w.String()
}

type SubmitResp struct {
	Header smgp.Header

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
	// (SP 根据请求和应答消息的 Sequence_Id 一致性就可得到 SMGP_Submit 消息的 Msg_Id)
	MsgID string

	// 1 字节，提交结果
	// 0:正确 1:消息结构错 2:命令字错 3:消息序号重复 4:消息长度错 5:资费代码错
	// 6:超过最大信息长 7:业务代码错 8:流量控制错 9~:其他错误
	Status uint32
}

func (s *SubmitResp) IDecode(data []byte) error {
	if len(data) < smgp.MinSMGPPduLength {
		return smgp.ErrInvalidPudLength
	}

	b := packet.NewPacketReader(data)
	defer b.Release()

	s.Header = smgp.ReadHeader(b)
	s.MsgID = hex.EncodeToString([]byte(b.ReadCStringNWithoutTrim(10)))
	s.Status = b.ReadUint32()

	return b.Error()
}

func (s *SubmitResp) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()

	smgp.WriteHeaderNoLength(s.Header, b)
	b.WriteFixedLenString(s.MsgID, 10)
	b.WriteUint32(s.Status)

	return b.BytesWithLength()
}

func (s *SubmitResp) SetSequenceID(id uint32) {
	s.Header.SequenceID = id
}

func (s *SubmitResp) GetSequenceID() uint32 {
	return s.Header.SequenceID
}

func (s *SubmitResp) GetCommand() sms.ICommander {
	return smgp.CommandSubmitResp
}

func (s *SubmitResp) GenEmptyResponse() sms.PDU {
	return nil
}

func (s *SubmitResp) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", s.Header)
	w.Write("MsgID", s.MsgID)
	w.Write("Status", s.Status)

	return w.String()
}
