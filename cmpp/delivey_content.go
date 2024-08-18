package cmpp

import "github.com/hujm2023/go-sms-protocol/packet"

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

	s.MsgID = r.ReadUint64()
	s.Stat = r.ReadCStringN(7)
	s.SubmitTime = r.ReadCStringN(10)
	s.DoneTime = r.ReadCStringN(10)
	s.DestTerminalID = r.ReadCStringN(21)
	s.SMSCSequence = r.ReadUint32()

	return r.Error()
}

func (s *SubPduDeliveryContent) SetSequenceID(_ uint32) {
}
