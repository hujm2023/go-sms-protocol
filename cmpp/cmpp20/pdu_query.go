package cmpp20

import (
	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/packet"
)

type PduQuery struct {
	cmpp.Header

	// 8 字节，时间。格式：YYYYMMDD(精确至日)
	Time string

	// 1 字节，查询类型：0 总数查询,1 按业务类型查询
	QueryType uint8

	// 10 字节，查询码：当 QueryType=0时，此项无效；当 QueryType=1时，此项填写业务类型 ServiceID
	QueryCode string

	// 8 字节，保留项
	Reserve string
}

func (p *PduQuery) IEncode() ([]byte, error) {
	p.TotalLength = MaxQueryLength
	b := packet.NewPacketWriter(MaxQueryLength)
	defer b.Release()

	b.WriteBytes(p.Header.Bytes())
	b.WriteFixedLenString(p.Time, 8)
	b.WriteUint8(p.QueryType)
	b.WriteFixedLenString(p.QueryCode, 10)
	b.WriteFixedLenString(p.Reserve, 8)

	return b.Bytes()
}

func (p *PduQuery) IDecode(data []byte) error {
	b := packet.NewPacketReader(data)
	defer b.Release()

	p.Header = cmpp.ReadHeader(b)
	p.Time = b.ReadCStringN(8)
	p.QueryType = b.ReadUint8()
	p.QueryCode = b.ReadCStringN(10)
	p.Reserve = b.ReadCStringN(8)

	return b.Error()
}

func (p *PduQuery) GetSequenceID() uint32 {
	return p.Header.SequenceID
}

func (p *PduQuery) SetSequenceID(sid uint32) {
	p.Header.SequenceID = sid
}

func (p *PduQuery) GetCommand() sms.ICommander {
	return cmpp.CommandQuery
}

func (p *PduQuery) GenEmptyResponse() sms.PDU {
	return &PduQueryResp{
		Header: cmpp.Header{
			CommandID:  cmpp.CommandQueryResp,
			SequenceID: p.GetSequenceID(),
		},
	}
}

func (p *PduQuery) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", p.Header)
	w.Write("Time", p.Time)
	w.Write("QueryType", p.QueryType)
	w.Write("QueryCode", p.QueryCode)
	w.Write("Reserve", p.Reserve)

	return w.String()
}

// -----------------------

type PduQueryResp struct {
	cmpp.Header

	// 8 字节，时间。格式：YYYYMMDD(精确至日)
	Time string

	// 1 字节，查询类型：0 总数查询, 1 按业务类型查询
	QueryType uint8

	// 10 字节，查询码：当 QueryType=0时，此项无效；当 QueryType=1时，此项填写业务类型 ServiceID
	QueryCode string

	// 4 字节，从 SP 接收信息总数
	MtTLMsg uint32

	// 4 字节，从 SP 接收用户总数
	MtTlUsr uint32

	// 4 字节，成功转发数量
	MtScs uint32

	// 4 字节，待转发数量
	MtWT uint32

	// 4 字节，转发失败数量
	MtFL uint32

	// 4 字节，向 SP 成功送达数量
	MoScs uint32

	// 4 字节，向 SP 待送达数量
	MoWT uint32

	// 4 字节，向 SP 送达失败数量
	MoFL uint32
}

func (p *PduQueryResp) IEncode() ([]byte, error) {
	p.TotalLength = MaxQueryRespLength
	b := packet.NewPacketWriter(MaxQueryRespLength)
	defer b.Release()

	b.WriteBytes(p.Header.Bytes())
	b.WriteFixedLenString(p.Time, 8)
	b.WriteUint8(p.QueryType)
	b.WriteFixedLenString(p.QueryCode, 10)
	b.WriteUint32(p.MtTLMsg)
	b.WriteUint32(p.MtTlUsr)
	b.WriteUint32(p.MtScs)
	b.WriteUint32(p.MtWT)
	b.WriteUint32(p.MtFL)
	b.WriteUint32(p.MtScs)
	b.WriteUint32(p.MtWT)
	b.WriteUint32(p.MtFL)

	return b.Bytes()
}

func (p *PduQueryResp) IDecode(data []byte) error {
	b := packet.NewPacketReader(data)
	defer b.Release()

	p.Header = cmpp.ReadHeader(b)
	p.Time = b.ReadCStringN(8)
	p.QueryType = b.ReadUint8()
	p.QueryCode = b.ReadCStringN(10)
	p.MtTLMsg = b.ReadUint32()
	p.MtTlUsr = b.ReadUint32()
	p.MtScs = b.ReadUint32()
	p.MtWT = b.ReadUint32()
	p.MtFL = b.ReadUint32()
	p.MtScs = b.ReadUint32()
	p.MtWT = b.ReadUint32()
	p.MtFL = b.ReadUint32()

	return b.Error()
}

func (p *PduQueryResp) GetSequenceID() uint32 {
	return p.Header.SequenceID
}

func (p *PduQueryResp) SetSequenceID(sid uint32) {
	p.Header.SequenceID = sid
}

func (p *PduQueryResp) GetCommand() sms.ICommander {
	return cmpp.CommandQueryResp
}

func (p *PduQueryResp) GenEmptyResponse() sms.PDU {
	return nil
}

func (p *PduQueryResp) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", p.Header)
	w.Write("Time", p.Time)
	w.Write("QueryType", p.QueryType)
	w.Write("QueryCode", p.QueryCode)
	w.Write("MtTLMsg", p.MtTLMsg)
	w.Write("MtTlUsr", p.MtTlUsr)
	w.Write("MtScs", p.MtScs)
	w.Write("MtWT", p.MtWT)
	w.Write("MtFL", p.MtFL)
	w.Write("MoScs", p.MoScs)
	w.Write("MoWT", p.MoWT)
	w.Write("MoFL", p.MoFL)

	return w.String()
}
