package sgip12

import (
	"strconv"

	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/packet"
	"github.com/hujm2023/go-sms-protocol/sgip"
)

type Report struct {
	sgip.Header

	// body
	// 12 字节 该命令所涉及的 Submit 或 deliver 命令的序列号
	SubmitSequence [3]uint32

	// 1 字节 Report 命令类型  0:对先前一条 Submit 命令的状态报告 1:对先前一条前转 Deliver 命令的状态报告
	ReportType uint8

	// 21 字节 接收短消息的手机号，手机号码前加“86”国别标 志
	UserNumber string

	// 1 字节 该命令所涉及的短消息的当前执行状态 0:发送成功  1:等待发送  2:发送失败
	State sgip.RespStatus

	// 1 字节 当 State=2 时为错误码值，否则为 0
	ErrorCode sgip.RespStatus

	// 8 字节 保留，扩展用
	Reserved string
}

func (p *Report) GetSubmitIdStr() string {
	return strconv.FormatUint(uint64(p.SubmitSequence[2]), 10)
}

func (p *Report) GetSubmitId() uint32 {
	return p.SubmitSequence[2]
}

func (p *Report) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()
	sgip.WriteHeaderNoLength(p.Header, b)
	b.WriteUint32(p.SubmitSequence[0])
	b.WriteUint32(p.SubmitSequence[1])
	b.WriteUint32(p.SubmitSequence[2])
	b.WriteUint8(p.ReportType)
	b.WriteFixedLenString(p.UserNumber, 21)
	b.WriteUint8(uint8(p.State))
	b.WriteUint8(uint8(p.ErrorCode))
	b.WriteFixedLenString(p.Reserved, 8)
	return b.BytesWithLength()
}

func (p *Report) IDecode(data []byte) error {
	if len(data) < sgip.MinSGIPPduLength {
		return sgip.ErrInvalidPudLength
	}
	b := packet.NewPacketReader(data)
	defer b.Release()
	p.Header = sgip.ReadHeader(b)
	p.SubmitSequence[0] = b.ReadUint32()
	p.SubmitSequence[1] = b.ReadUint32()
	p.SubmitSequence[2] = b.ReadUint32()
	p.ReportType = b.ReadUint8()
	p.UserNumber = b.ReadCStringN(21)
	p.State = sgip.RespStatus(b.ReadUint8())
	p.ErrorCode = sgip.RespStatus(b.ReadUint8())
	p.Reserved = b.ReadCStringN(8)
	return b.Error()
}

func (p *Report) SetSequenceID(id uint32) {
	p.Header.Sequence[2] = id
}

func (r *Report) GetSequenceID() uint32 {
	return r.Header.Sequence[2]
}

func (r *Report) GetCommand() sms.ICommander {
	return sgip.SGIP_REPORT
}

func (r *Report) GenEmptyResponse() sms.PDU {
	return &ReportResp{
		Header: sgip.NewHeader(0, sgip.SGIP_REPORT_REP, r.Header.Sequence[0], r.GetSequenceID()),
	}
}

func (r *Report) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", r.Header)
	w.Write("SubmitSequence", r.SubmitSequence)
	w.Write("ReportType", r.ReportType)
	w.Write("UserNumber", r.UserNumber)
	w.Write("State", r.State)
	w.Write("ErrorCode", r.ErrorCode)
	w.Write("Reserved", r.Reserved)

	return w.String()
}

type ReportResp struct {
	sgip.Header

	// 1 字节 Report命令是否成功接收。 0:接收成功 其它:错误码
	Result sgip.RespStatus

	// 8 字节 保留，扩展用
	Reserved string
}

func (p *ReportResp) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()
	sgip.WriteHeaderNoLength(p.Header, b)
	b.WriteUint8(uint8(p.Result))
	b.WriteFixedLenString(p.Reserved, 8)
	return b.BytesWithLength()
}

func (p *ReportResp) IDecode(data []byte) error {
	b := packet.NewPacketReader(data)
	defer b.Release()
	p.Header = sgip.ReadHeader(b)
	p.Result = sgip.RespStatus(b.ReadUint8())
	p.Reserved = b.ReadCStringN(8)
	return b.Error()
}

func (p *ReportResp) SetSequenceID(id uint32) {
	p.Header.Sequence[2] = id
}

func (r *ReportResp) GetSequenceID() uint32 {
	return r.Header.Sequence[2]
}

func (r *ReportResp) GetCommand() sms.ICommander {
	return sgip.SGIP_REPORT_REP
}

func (r *ReportResp) GenEmptyResponse() sms.PDU {
	return nil
}

func (r *ReportResp) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", r.Header)
	w.Write("Result", r.Result)
	w.Write("Reserved", r.Reserved)

	return w.String()
}
