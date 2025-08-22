package smgp30

import (
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/packet"
	"github.com/hujm2023/go-sms-protocol/smgp"
)

type Deliver struct {
	smgp.Header

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
	MsgID string

	// 是否为状态报告
	IsReport uint8

	// 短消息格式
	MsgFormat uint8

	// 短消息接收时间
	RecvTime string

	// 短消息发送号码
	SrcTermID string

	// 短消息接受号码
	DestTermID string

	// 短消息长度
	MsgLength uint8

	// 短消息内容
	MsgContent []byte

	// 保留
	Reserve string

	Options smgp.Options
}

func (d *Deliver) IDecode(data []byte) error {
	if len(data) < smgp.MinSMGPPduLength {
		return smgp.ErrInvalidPudLength
	}

	b := packet.NewPacketReader(data)
	defer b.Release()

	d.Header = smgp.ReadHeader(b)
	d.MsgID = hex.EncodeToString([]byte(b.ReadCStringNWithoutTrim(10)))
	d.IsReport = b.ReadUint8()
	d.MsgFormat = b.ReadUint8()
	d.RecvTime = b.ReadCStringN(14)
	d.SrcTermID = b.ReadCStringN(21)
	d.DestTermID = b.ReadCStringN(21)
	d.MsgLength = b.ReadUint8()
	d.MsgContent = b.ReadNBytes(int(d.MsgLength))
	d.Reserve = b.ReadCStringN(8)
	d.Options = smgp.ReadOptions(b)

	return b.Error()
}

func (d *Deliver) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()

	smgp.WriteHeaderNoLength(d.Header, b)
	b.WriteFixedLenString(d.MsgID, 10)
	b.WriteUint8(d.IsReport)
	b.WriteUint8(d.MsgFormat)
	b.WriteFixedLenString(d.RecvTime, 14)
	b.WriteFixedLenString(d.SrcTermID, 21)
	b.WriteFixedLenString(d.DestTermID, 21)
	b.WriteUint8(d.MsgLength)
	b.WriteBytes(d.MsgContent)
	b.WriteFixedLenString(d.Reserve, 8)
	b.WriteBytes(d.Options.Serialize())

	return b.BytesWithLength()
}

func (d *Deliver) SetSequenceID(id uint32) {
	d.Header.SequenceID = id
}

func (d *Deliver) GetSequenceID() uint32 {
	return d.Header.SequenceID
}

func (d *Deliver) GenerateResponseHeader() *DeliverResp {
	resp := &DeliverResp{
		Header: smgp.NewHeader(smgp.MaxDeliverRespLength, smgp.CommandDeliverResp, d.GetSequenceID()),
	}
	return resp
}

func (d *Deliver) GetCommand() sms.ICommander {
	return smgp.CommandDeliver
}

func (d *Deliver) GenEmptyResponse() sms.PDU {
	return &DeliverResp{
		Header: smgp.NewHeader(smgp.MaxDeliverRespLength, smgp.CommandDeliverResp, d.GetSequenceID()),
	}
}

func (d *Deliver) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", d.Header)
	w.Write("MsgID", d.MsgID)
	w.Write("IsReport", d.IsReport)
	w.Write("MsgFormat", d.MsgFormat)
	w.Write("RecvTime", d.RecvTime)
	w.Write("SrcTermID", d.SrcTermID)
	w.Write("DestTermID", d.DestTermID)
	w.Write("MsgLength", d.MsgLength)
	w.WriteWithBytes("MsgContent", d.MsgContent)
	w.Write("Reserve", d.Reserve)
	w.OmitWrite("Options", d.Options.String())

	return w.String()
}

type DeliverResp struct {
	Header smgp.Header

	MsgID string

	// 4 字节, 回执结果
	// 0:正确
	// 1:消息结构错 2:命令字错 3:消息序号重复 4:消息长度错 5:资费代码错
	// 6:超过最大信息长 7:业务代码错 8: 流量控制错 9~ :其他错误
	Result smgp.Status
}

func (d *DeliverResp) IDecode(data []byte) error {
	if len(data) < smgp.MinSMGPPduLength {
		return smgp.ErrInvalidPudLength
	}
	b := packet.NewPacketReader(data)
	defer b.Release()

	d.Header = smgp.ReadHeader(b)
	d.MsgID = hex.EncodeToString([]byte(b.ReadCStringNWithoutTrim(10)))
	d.Result = smgp.Status(b.ReadUint32())

	return b.Error()
}

func (d *DeliverResp) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()

	smgp.WriteHeaderNoLength(d.Header, b)
	msgID, _ := hex.DecodeString(d.MsgID)
	b.WriteFixedLenString(string(msgID), 10)
	b.WriteUint32(d.Result.Data())

	return b.BytesWithLength()
}

func (d *DeliverResp) SetSequenceID(id uint32) {
	d.Header.SequenceID = id
}

func (d *DeliverResp) GetSequenceID() uint32 {
	return d.Header.SequenceID
}

func (d *DeliverResp) GetCommand() sms.ICommander {
	return smgp.CommandDeliverResp
}

func (d *DeliverResp) GenEmptyResponse() sms.PDU {
	return nil
}

func (d *DeliverResp) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", d.Header)
	w.Write("MsgID", d.MsgID)
	w.Write("Result", d.Result)

	return w.String()
}

// DeliveryReceipt is the model representation of short_message for SMPP PDU delivery_sm.
// SMPP provides for return of an SMSC delivery receipt via the deliver_sm or data_sm PDU,which indicates the delivery status of the message.
type DeliveryReceipt struct {
	ID       string // id，10，C-Octet String (Decimal)，The message ID allocated to the message by the SMSC when originally submitted.
	Sub      string // sub， 3， C-Octet String (Decimal)，Number of short messages originally submitted. This is only relevant when the original message was submitted to a distribution list.The value is padded with leading zeros if necessary.
	Dlvrd    string // dlvrd， 3， C-Octet Fixed Length String (Decimal)， Number of short messages delivered. This is only relevant where the original message was submitted to a distribution list.The value is padded with leading zeros if necessary.
	SubDate  string // submit date 10，C-Octet Fixed Length String The time and date at which the short message was submitted. In the case of a message which has been replaced, this is the date that the original message was replaced.The format is as follows: YYMMDDhhmm where: YY = last two digits of the year (00-99) MM = month (01-12) ，DD= day (01-31) hh = hour (00-23) mm = minute (00-59)
	DoneDate string // done date，10，C-Octet Fixed Length String，The time and date at which the short message reached it’s final state. The format is the same as for the submit date.
	Stat     string // stat，7，C-Octet Fixed Length String，The final status of the message. For settings for this field see Table B-2.
	Err      string // err，3，C-Octet Fixed Length String，Where appropriate this may hold a Network specific error code or an SMSC error code for the attempted delivery of the message. These errors are Network or SMSC specific and are not included here.
	Text     string // text，20，Octet String，The first 20 characters of the short message.
}

func (d DeliveryReceipt) Valid() bool {
	// id 和 stat 都有值时才有效
	if d.ID != "" && d.Stat != "" {
		return true
	}
	return false
}

// ExtractDeliveryReceipt 将short_message字符串提取为 DeliveryReceipt 结构体.
func (d *DeliveryReceipt) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	buf.WriteString("id:")
	id, err := hex.DecodeString(d.ID)
	if err != nil {
		return nil, err
	}
	buf.WriteBytes(id)
	buf.WriteString(" sub:")
	buf.WriteString(d.Sub)
	buf.WriteString(" dlvrd:")
	buf.WriteString(d.Dlvrd)
	buf.WriteString(" submit date:")
	buf.WriteString(d.SubDate)
	buf.WriteString(" done date:")
	buf.WriteString(d.DoneDate)
	buf.WriteString(" stat:")
	buf.WriteString(d.Stat)
	buf.WriteString(" err:")
	buf.WriteString(d.Err)
	buf.WriteString(" text:")
	buf.WriteString(d.Text)
	return buf.Bytes()
}

func (d *DeliveryReceipt) IDecode(data []byte) error {
	// 去掉最后的 0x00
	s := strings.TrimRight(string(data), "\x00")
	d.ID = findSMGPIDValue(s)
	d.Sub = findSubValue(s, "sub", "Sub", 3)
	d.Dlvrd = findSubValue(s, "dlvrd", "Dlvrd", 3)
	d.SubDate = findSubValue(s, "submit date", "Submit_Date", 10)
	d.DoneDate = findSubValue(s, "done date", "Done_Date", 10)
	d.Stat = findSubValue(s, "stat", "Stat", 7)
	d.Err = findSubValue(s, "err", "Err", 3)
	d.Text = findSubValue(s, "text", "Text", 20)
	return nil
}

func ExtractDeliveryReceipt(s string) (d DeliveryReceipt, err error) {
	d.ID = findSMGPIDValue(s)
	d.Sub = findSubValue(s, "sub", "Sub", 3)
	d.Dlvrd = findSubValue(s, "dlvrd", "Dlvrd", 3)
	d.SubDate = findSubValue(s, "submit date", "Submit_Date", 10)
	d.DoneDate = findSubValue(s, "done date", "Done_Date", 10)
	d.Stat = findSubValue(s, "stat", "Stat", 7)
	d.Err = findSubValue(s, "err", "Err", 3)
	d.Text = findSubValue(s, "text", "Text", 20)
	return
}

func findSubValue(s string, sub, backup string, maxSize int) (value string) {
	// maxSize = 0 // 先不校验

	sub = sub + ":"
	n := strings.Index(s, sub)
	if n == -1 {
		if backup == "" {
			return
		}
		// 用backup再找一遍
		n = strings.Index(s, backup)
		if n == -1 {
			return
		}
	}

	start := n + len(sub)
	// 当前 key 后面的下一个空格
	spaceIdx := strings.Index(s[start:], " ")

	// 后面再无空格，说明当前 key 是最后一个，直接返回即可
	if spaceIdx == -1 {
		value = s[start:]
	} else {
		// 空格之前的就是我们要找的 value
		value = s[start : start+spaceIdx]
	}

	if maxSize > 0 && len(value) > maxSize {
		value = value[:maxSize]
	}

	return
}

// 采用bcd编码时可能会出现值为32的字节即空格，所以不用空格判断 直接用长度
func findSMGPIDValue(s string) (value string) {
	l := 10
	sub := "id:"
	n := strings.Index(s, sub)
	if n == -1 {
		return ""
	}
	start := n + len(sub)
	if len(s) >= start+l {
		value = s[start : start+l]
	} else {
		return ""
	}

	return hex.EncodeToString([]byte(value))
}

// ExtractDeliveryReceipt1 将short_message字符串提取为 DeliveryReceipt 结构体.
// Deprecated. 性能不如 ExtractDeliveryReceipt,且不能处理key顺序异常、缺少 key 的场景
func ExtractDeliveryReceipt1(s string) (d DeliveryReceipt, err error) {
	_, err = fmt.Sscanf(s, "id:%s sub:%s dlvrd:%s submit date:%s done date:%s stat:%s err:%s text:%s",
		&d.ID, &d.Sub, &d.Dlvrd, &d.SubDate, &d.DoneDate, &d.Stat, &d.Err, &d.Text)
	if err != nil && err == io.EOF {
		return d, nil
	}
	return
}
