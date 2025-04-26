package cmpp30

import (
	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/packet"
)

// Query represents a CMPP 3.0 Query PDU.
// It is used by the SP to query the statistics of messages sent on a specific day.
type Query struct {
	cmpp.Header

	// Time is the time in YYYYMMDD format (8 bytes).
	Time string

	// QueryType indicates the query type (1 byte): 0 for total count, 1 for query by service type.
	QueryType uint8

	// QueryCode is the query code (10 bytes). Invalid when QueryType=0; ServiceID when QueryType=1.
	QueryCode string

	// Reserve is a reserved field (8 bytes).
	Reserve string
}

// IDecode decodes the byte slice into a Query PDU.
// IDecode decodes the byte slice into a Query PDU.
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

// IEncode encodes the Query PDU into a byte slice.
// IEncode encodes the Query PDU into a byte slice.
func (q *Query) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()

	cmpp.WriteHeaderNoLength(q.Header, b)
	b.WriteFixedLenString(q.Time, 8)
	b.WriteUint8(q.QueryType)
	b.WriteFixedLenString(q.QueryCode, 10)
	b.WriteFixedLenString(q.Reserve, 8)

	return b.BytesWithLength()
}

// SetSequenceID sets the sequence ID of the PDU.
// SetSequenceID sets the sequence ID of the PDU.
func (q *Query) SetSequenceID(id uint32) {
	q.Header.SequenceID = id
}

// GetSequenceID returns the sequence ID of the PDU.
// GetSequenceID returns the sequence ID of the PDU.
func (q *Query) GetSequenceID() uint32 {
	return q.Header.SequenceID
}

// GetCommand returns the command ID of the PDU.
// GetCommand returns the command ID of the PDU.
func (q *Query) GetCommand() sms.ICommander {
	return cmpp.CommandQuery
}

// GenEmptyResponse generates an empty response PDU for the Query.
// GenEmptyResponse generates an empty response PDU for the Query.
func (q *Query) GenEmptyResponse() sms.PDU {
	return &QueryResp{
		Header: cmpp.Header{
			CommandID:  cmpp.CommandQueryResp,
			SequenceID: q.GetSequenceID(),
		},
	}
}

// String returns a string representation of the Query PDU.
// String returns a string representation of the Query PDU.
func (q *Query) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", q.Header)
	w.Write("Time", q.Time)
	w.Write("QueryType", q.QueryType)
	w.Write("QueryCode", q.QueryCode)
	w.Write("Reserve", q.Reserve)

	return w.String()
}

// QueryResp represents a CMPP 3.0 QueryResp PDU.
// It is the response to a Query PDU.
type QueryResp struct {
	cmpp.Header

	// Time is the time in YYYYMMDD format (8 bytes).
	Time string

	// QueryType indicates the query type (1 byte): 0 for total count, 1 for query by service type.
	QueryType uint8

	// QueryCode is the query code (10 bytes). Invalid when QueryType=0; ServiceID when QueryType=1.
	QueryCode string

	// MtTLMsg is the total number of MT messages received from SP (4 bytes).
	MtTLMsg uint32

	// MtTlUsr is the total number of MT users received from SP (4 bytes).
	MtTlUsr uint32

	// MtScs is the number of successfully forwarded MT messages (4 bytes).
	MtScs uint32

	// MtWT is the number of MT messages waiting to be forwarded (4 bytes).
	MtWT uint32

	// MtFL is the number of failed MT messages (4 bytes).
	MtFL uint32

	// MoScs is the number of MO messages successfully delivered to SP (4 bytes).
	MoScs uint32

	// MoWT is the number of MO messages waiting to be delivered to SP (4 bytes).
	MoWT uint32

	// MoFL is the number of MO messages failed to be delivered to SP (4 bytes).
	MoFL uint32
}

// IDecode decodes the byte slice into a QueryResp PDU.
// IDecode decodes the byte slice into a QueryResp PDU.
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

// IEncode encodes the QueryResp PDU into a byte slice.
// IEncode encodes the QueryResp PDU into a byte slice.
func (q *QueryResp) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()

	cmpp.WriteHeaderNoLength(q.Header, b)
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

// SetSequenceID sets the sequence ID of the PDU.
// SetSequenceID sets the sequence ID of the PDU.
func (q *QueryResp) SetSequenceID(id uint32) {
	q.Header.SequenceID = id
}

// GetSequenceID returns the sequence ID of the PDU.
// GetSequenceID returns the sequence ID of the PDU.
func (q *QueryResp) GetSequenceID() uint32 {
	return q.Header.SequenceID
}

// GetCommand returns the command ID of the PDU.
// GetCommand returns the command ID of the PDU.
func (q *QueryResp) GetCommand() sms.ICommander {
	return cmpp.CommandQueryResp
}

// GenEmptyResponse generates an empty response PDU (nil for QueryResp).
// GenEmptyResponse generates an empty response PDU (nil for QueryResp).
func (q *QueryResp) GenEmptyResponse() sms.PDU {
	return nil
}

// String returns a string representation of the QueryResp PDU.
// String returns a string representation of the QueryResp PDU.
func (q *QueryResp) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", q.Header)
	w.Write("Time", q.Time)
	w.Write("QueryType", q.QueryType)
	w.Write("QueryCode", q.QueryCode)
	w.Write("MtTLMsg", q.MtTLMsg)
	w.Write("MtTlUsr", q.MtTlUsr)
	w.Write("MtScs", q.MtScs)
	w.Write("MtWT", q.MtWT)
	w.Write("MtFL", q.MtFL)
	w.Write("MoScs", q.MoScs)
	w.Write("MoWT", q.MoWT)
	w.Write("MoFL", q.MoFL)

	return w.String()
}
