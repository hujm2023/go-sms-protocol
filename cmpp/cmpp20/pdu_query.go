package cmpp20

import (
	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/packet"
)

// PduQuery represents a CMPP Query PDU.
// It is used to query the status of the ISMG.
type PduQuery struct {
	cmpp.Header

	// Time is the time (8 bytes, YYYYMMDD format).
	Time string

	// QueryType is the query type (1 byte): 0=Total query, 1=Query by service type.
	QueryType uint8

	// QueryCode is the query code (10 bytes): Invalid when QueryType=0; ServiceID when QueryType=1.
	QueryCode string

	// Reserve is a reserved field (8 bytes).
	Reserve string
}

// IEncode encodes the PduQuery PDU into a byte slice.
func (p *PduQuery) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter(0)
	defer b.Release()

	cmpp.WriteHeaderNoLength(p.Header, b)
	b.WriteFixedLenString(p.Time, 8)
	b.WriteUint8(p.QueryType)
	b.WriteFixedLenString(p.QueryCode, 10)
	b.WriteFixedLenString(p.Reserve, 8)

	return b.BytesWithLength()
}

// IDecode decodes the byte slice into a PduQuery PDU.
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

// GetSequenceID returns the sequence ID of the PDU.
func (p *PduQuery) GetSequenceID() uint32 {
	return p.Header.SequenceID
}

// SetSequenceID sets the sequence ID of the PDU.
func (p *PduQuery) SetSequenceID(sid uint32) {
	p.Header.SequenceID = sid
}

// GetCommand returns the command ID of the PDU.
func (p *PduQuery) GetCommand() sms.ICommander {
	return cmpp.CommandQuery
}

// GenEmptyResponse generates an empty response PDU for the PduQuery.
func (p *PduQuery) GenEmptyResponse() sms.PDU {
	return &PduQueryResp{
		Header: cmpp.Header{
			CommandID:  cmpp.CommandQueryResp,
			SequenceID: p.GetSequenceID(),
		},
	}
}

// String returns a string representation of the PduQuery PDU.
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

// PduQueryResp represents a CMPP QueryResp PDU.
// It is the response to a PduQuery.
type PduQueryResp struct {
	cmpp.Header

	// Time is the time (8 bytes, YYYYMMDD format).
	Time string

	// QueryType is the query type (1 byte): 0=Total query, 1=Query by service type.
	QueryType uint8

	// QueryCode is the query code (10 bytes): Invalid when QueryType=0; ServiceID when QueryType=1.
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

	// MoScs is the number of successfully delivered MO messages to SP (4 bytes).
	MoScs uint32

	// MoWT is the number of MO messages waiting to be delivered to SP (4 bytes).
	MoWT uint32

	// MoFL is the number of failed MO messages to SP (4 bytes).
	MoFL uint32
}

// IEncode encodes the PduQueryResp PDU into a byte slice.
func (p *PduQueryResp) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter(0)
	defer b.Release()

	cmpp.WriteHeaderNoLength(p.Header, b)
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

	return b.BytesWithLength()
}

// IDecode decodes the byte slice into a PduQueryResp PDU.
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

// GetSequenceID returns the sequence ID of the PDU.
func (p *PduQueryResp) GetSequenceID() uint32 {
	return p.Header.SequenceID
}

// SetSequenceID sets the sequence ID of the PDU.
func (p *PduQueryResp) SetSequenceID(sid uint32) {
	p.Header.SequenceID = sid
}

// GetCommand returns the command ID of the PDU.
func (p *PduQueryResp) GetCommand() sms.ICommander {
	return cmpp.CommandQueryResp
}

// GenEmptyResponse generates an empty response PDU (nil for QueryResp).
func (p *PduQueryResp) GenEmptyResponse() sms.PDU {
	return nil
}

// String returns a string representation of the PduQueryResp PDU.
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
