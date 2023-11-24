package cmpp30

import (
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/packet"
)

type Query struct {
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

func (q *Query) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return cmpp.ErrInvalidPudLength
	}
	b := packet.NewPacketReader(data)
	defer b.Release()

	q.Header = cmpp.ReadHeader(b)
	q.Time = b.ReadCStringN(8)
	q.QueryType = b.ReadUint8()
	q.QueryCode = b.ReadCStringN(10)
	q.Reserve = b.ReadCStringN(8)

	return b.Error()
}

func (q *Query) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()

	b.WriteUint32(uint32(q.Header.CommandID))
	b.WriteUint32(q.Header.SequenceID)
	b.WriteFixedLenString(q.Time, 8)
	b.WriteUint8(q.QueryType)
	b.WriteFixedLenString(q.QueryCode, 10)
	b.WriteFixedLenString(q.Reserve, 8)

	return b.BytesWithLength()
}

func (q *Query) SetSequenceID(id uint32) {
	q.Header.SequenceID = id
}

type QueryResp struct {
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

func (q *QueryResp) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return cmpp.ErrInvalidPudLength
	}
	b := packet.NewPacketReader(data)
	defer b.Release()

	q.Header = cmpp.ReadHeader(b)
	q.Time = b.ReadCStringN(8)
	q.QueryType = b.ReadUint8()
	q.QueryCode = b.ReadCStringN(10)
	q.MtTLMsg = b.ReadUint32()
	q.MtTlUsr = b.ReadUint32()
	q.MtScs = b.ReadUint32()
	q.MtWT = b.ReadUint32()
	q.MtFL = b.ReadUint32()
	q.MoScs = b.ReadUint32()
	q.MoWT = b.ReadUint32()
	q.MoFL = b.ReadUint32()

	return b.Error()
}

func (q *QueryResp) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()

	b.WriteUint32(uint32(q.Header.CommandID))
	b.WriteUint32(q.Header.SequenceID)
	b.WriteFixedLenString(q.Time, 8)
	b.WriteUint8(q.QueryType)
	b.WriteFixedLenString(q.QueryCode, 10)
	b.WriteUint32(q.MtTLMsg)
	b.WriteUint32(q.MtTlUsr)
	b.WriteUint32(q.MtScs)
	b.WriteUint32(q.MtWT)
	b.WriteUint32(q.MtFL)
	b.WriteUint32(q.MoScs)
	b.WriteUint32(q.MoWT)
	b.WriteUint32(q.MoFL)

	return b.BytesWithLength()
}

func (q *QueryResp) SetSequenceID(id uint32) {
	q.Header.SequenceID = id
}
