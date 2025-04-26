package cmpp20

import (
	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/packet"
)

// PduDeliver represents a CMPP Deliver PDU.
// It is used by the ISMG to deliver a short message to the SP, or to send a status report.
type PduDeliver struct {
	cmpp.Header

	/* MsgID is the message identifier (8 bytes, uint64):
	(1) Time (MMDDHHMMSS format): bits 64-39
	    - bits 64-61: Month (binary)
	    - bits 60-56: Day (binary)
	    - bits 55-51: Hour (binary)
	    - bits 50-45: Minute (binary)
	    - bits 44-39: Second (binary)
	(2) SMS Gateway Code: bits 38-17 (integer representation)
	(3) Sequence Number: bits 16-1 (incrementing, wraps around)
	Left-pad with zeros if necessary, right-aligned.
	*/
	MsgID uint64

	// DestID is the destination number (21 bytes).
	// SP's service code (usually 4-6 digits) or a long number prefixed with the service code.
	// This is the recipient number for the mobile user's short message.
	DestID string

	// ServiceID is the service type (10 bytes, combination of digits, letters, and symbols).
	ServiceID string

	// TpPID is the GSM protocol type (1 byte). See GSM 03.40 section 9.2.3.9.
	TpPID uint8

	// TpUDHI is the GSM protocol type (1 byte). See GSM 03.40 section 9.2.3.23 (only 1 bit used, right-aligned).
	TpUDHI uint8

	// MsgFmt is the message format (1 byte): 0=ASCII, 3=SMS Write Card, 4=Binary, 8=UCS2, 15=GB Hanzi.
	MsgFmt uint8

	// SrcTerminalID is the source terminal MSISDN (21 bytes). For status reports, this is the destination terminal number from the PduSubmit.
	SrcTerminalID string `safe_json:"-"`

	// RegisteredDeliver indicates if it's a status report (1 byte): 0=Not a status report (MO), 1=Status report (MT).
	RegisteredDeliver uint8

	// MsgLength is the message length (1 byte).
	MsgLength uint8

	// MsgContent is the message content (MsgLength bytes).
	MsgContent string

	// Reserved is a reserved field (8 bytes).
	Reserved string
}

// IEncode encodes the PduDeliver PDU into a byte slice.
func (p *PduDeliver) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter(0)
	defer b.Release()

	cmpp.WriteHeaderNoLength(p.Header, b)
	b.WriteUint64(p.MsgID)
	b.WriteFixedLenString(p.DestID, 21)
	b.WriteFixedLenString(p.ServiceID, 10)
	b.WriteUint8(p.TpPID)
	b.WriteUint8(p.TpUDHI)
	b.WriteUint8(p.MsgFmt)
	b.WriteFixedLenString(p.SrcTerminalID, 21)
	b.WriteUint8(p.RegisteredDeliver)
	b.WriteUint8(p.MsgLength)
	b.WriteString(p.MsgContent)
	b.WriteFixedLenString(p.Reserved, 8)

	return b.BytesWithLength()
}

// IDecode decodes the byte slice into a PduDeliver PDU.
func (p *PduDeliver) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return ErrInvalidPudLength
	}

	b := packet.NewPacketReader(data)
	defer b.Release()

	p.Header = cmpp.ReadHeader(b)
	p.MsgID = b.ReadUint64()
	p.DestID = b.ReadCStringN(21)
	p.ServiceID = b.ReadCStringN(10)
	p.TpPID = b.ReadUint8()
	p.TpUDHI = b.ReadUint8()
	p.MsgFmt = b.ReadUint8()
	p.SrcTerminalID = b.ReadCStringN(21)
	p.RegisteredDeliver = b.ReadUint8()
	p.MsgLength = b.ReadUint8()
	p.MsgContent = string(b.ReadNBytes(int(p.MsgLength)))
	p.Reserved = b.ReadCStringN(8)

	return b.Error()
}

// GetSequenceID returns the sequence ID of the PDU.
func (p *PduDeliver) GetSequenceID() uint32 {
	return p.Header.SequenceID
}

// SetSequenceID sets the sequence ID of the PDU.
func (p *PduDeliver) SetSequenceID(sid uint32) {
	p.Header.SequenceID = sid
}

// GetCommand returns the command ID of the PDU.
func (p *PduDeliver) GetCommand() sms.ICommander {
	return cmpp.CommandDeliver
}

// GenEmptyResponse generates an empty response PDU for the PduDeliver.
func (p *PduDeliver) GenEmptyResponse() sms.PDU {
	return &PduDeliverResp{
		Header: cmpp.Header{
			CommandID:  cmpp.CommandDeliverResp,
			SequenceID: p.GetSequenceID(),
		},
	}
}

// String returns a string representation of the PduDeliver PDU.
func (p *PduDeliver) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", p.Header)
	w.Write("MsgID", p.MsgID)
	w.Write("DestID", p.DestID)
	w.Write("ServiceID", p.ServiceID)
	w.Write("TpPID", p.TpPID)
	w.Write("TpUDHI", p.TpUDHI)
	w.Write("MsgFmt", p.MsgFmt)
	w.Write("SrcTerminalID", p.SrcTerminalID)
	w.Write("RegisteredDeliver", p.RegisteredDeliver)
	w.Write("MsgLength", p.MsgLength)
	w.WriteWithBytes("MsgContent", p.MsgContent)
	w.Write("Reserved", p.Reserved)

	return w.String()
}

// --------------------------------------

// PduDeliverResp represents a CMPP DeliverResp PDU.
// It is the response to a PduDeliver PDU.
type PduDeliverResp struct {
	cmpp.Header

	// MsgID is the MsgID from the corresponding PduDeliver (8 bytes).
	MsgID uint64

	// Result indicates the outcome (1 byte):
	// 0: Correct
	// 1: Invalid message structure, 2: Invalid command ID, 3: Duplicate sequence number,
	// 4: Invalid message length, 5: Invalid fee code, 6: Exceeded max message length,
	// 7: Invalid service ID, 8: Flow control error, 9+: Other errors.
	Result uint8
}

// IEncode encodes the PduDeliverResp PDU into a byte slice.
func (pr *PduDeliverResp) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter(0)
	defer b.Release()

	cmpp.WriteHeaderNoLength(pr.Header, b)
	b.WriteUint64(pr.MsgID)
	b.WriteUint8(pr.Result)

	return b.BytesWithLength()
}

// IDecode decodes the byte slice into a PduDeliverResp PDU.
func (pr *PduDeliverResp) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return ErrInvalidPudLength
	}

	b := packet.NewPacketReader(data)
	defer b.Release()

	pr.Header = cmpp.ReadHeader(b)
	pr.MsgID = b.ReadUint64()
	pr.Result = b.ReadUint8()

	return b.Error()
}

// GetSequenceID returns the sequence ID of the PDU.
func (pr *PduDeliverResp) GetSequenceID() uint32 {
	return pr.Header.SequenceID
}

// SetSequenceID sets the sequence ID of the PDU.
func (pr *PduDeliverResp) SetSequenceID(sid uint32) {
	pr.Header.SequenceID = sid
}

// GetCommand returns the command ID of the PDU.
func (p *PduDeliverResp) GetCommand() sms.ICommander {
	return cmpp.CommandDeliverResp
}

// GenEmptyResponse generates an empty response PDU (nil for DeliverResp).
func (p *PduDeliverResp) GenEmptyResponse() sms.PDU {
	return nil
}

// String returns a string representation of the PduDeliverResp PDU.
func (p *PduDeliverResp) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", p.Header)
	w.Write("MsgID", p.MsgID)
	w.Write("Result", p.Result)

	return w.String()
}
