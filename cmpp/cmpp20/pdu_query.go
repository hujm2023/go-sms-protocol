package cmpp20

import (
	"github.com/hujm2023/go-sms-protocol"
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

	b.WriteBytes(p.GetHeader().Bytes())
	b.WriteFixedLenString(p.Time, 8)
	b.WriteUint8(p.QueryType)
	b.WriteFixedLenString(p.QueryCode, 10)
	b.WriteFixedLenString(p.Reserve, 8)

	return b.Bytes()
}

func (p *PduQuery) IDecode(data []byte) error {
	b := packet.NewPacketReader(data)
	defer b.Release()

	p.Header = b.ReadHeader()
	p.Time = b.ReadCStringN(8)
	b.ReadNumeric(&p.QueryType)
	p.QueryCode = b.ReadCStringN(10)
	p.Reserve = b.ReadCStringN(8)

	return b.Error()
}

func (p *PduQuery) GetHeader() cmpp.Header {
	return p.Header
}

func (p *PduQuery) GetSequenceID() uint32 {
	return p.GetHeader().SequenceID
}

func (p *PduQuery) GetCommandID() cmpp.CommandID {
	return cmpp.CommandQuery
}

func (p *PduQuery) GenerateResponseHeader() protocol.PDU {
	resp := &PduQueryResp{
		Header: cmpp.NewHeader(MaxQueryRespLength, cmpp.CommandQueryResp, p.GetSequenceID()),
	}
	return resp
}

func (p *PduQuery) MaxLength() uint32 {
	return MaxQueryLength
}

func (p *PduQuery) SetSequenceID(sid uint32) {
	p.Header.SequenceID = sid
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

	p.Header = b.ReadHeader()
	p.Time = b.ReadCStringN(8)
	b.ReadNumeric(&p.QueryType)
	p.QueryCode = b.ReadCStringN(10)
	b.ReadNumeric(&p.MtTLMsg)
	b.ReadNumeric(&p.MtTlUsr)
	b.ReadNumeric(&p.MtScs)
	b.ReadNumeric(&p.MtWT)
	b.ReadNumeric(&p.MtFL)
	b.ReadNumeric(&p.MtScs)
	b.ReadNumeric(&p.MtWT)
	b.ReadNumeric(&p.MtFL)

	return b.Error()
}

func (p *PduQueryResp) GetHeader() cmpp.Header {
	return p.Header
}

func (p *PduQueryResp) GetSequenceID() uint32 {
	return p.GetHeader().SequenceID
}

func (p *PduQueryResp) GetCommandID() cmpp.CommandID {
	return cmpp.CommandQueryResp
}

func (p *PduQueryResp) GenerateResponseHeader() protocol.PDU {
	return nil
}

func (p *PduQueryResp) MaxLength() uint32 {
	return MaxQueryRespLength
}

func (p *PduQueryResp) SetSequenceID(sid uint32) {
	p.Header.SequenceID = sid
}
