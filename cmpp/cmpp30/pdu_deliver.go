package cmpp30

import (
	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/packet"
)

// Deliver represents a CMPP 3.0 Deliver PDU.
// It is used by the ISMG to deliver a short message (including status reports) to the SP.
type Deliver struct {
	cmpp.Header

	/* MsgID is the message identifier (8 bytes), a 64-bit integer:
	(1) Time (MMDDHHMMSS format): bits 64-39
	    - bits 64-61: Month (binary)
	    - bits 60-56: Day (binary)
	    - bits 55-51: Hour (binary)
	    - bits 50-45: Minute (binary)
	    - bits 44-39: Second (binary)
	(2) SMS Gateway Code: bits 38-17 (integer representation of the gateway code)
	(3) Sequence Number: bits 16-1 (sequentially increasing, wraps around)
	Left-pad with zeros if necessary, right-aligned.
	*/
	MsgID uint64

	// DestID is the destination number (21 bytes).
	// SP's service code (usually 4-6 digits) or a long number prefixed with the service code.
	// This is the called number for the mobile user's short message.
	DestID string

	// ServiceID is the service type (10 bytes), a combination of digits, letters, and symbols.
	ServiceID string

	// TpPID is the GSM protocol type (1 byte). See GSM 03.40 section 9.2.3.9.
	TpPID uint8

	// TpUDHI is the GSM protocol type (1 byte). See GSM 03.40 section 9.2.3.23 (only 1 bit used, right-aligned).
	TpUDHI uint8

	// MsgFmt is the message format (1 byte): 0=ASCII, 3=SMS Write Card, 4=Binary, 8=UCS2, 15=GB Hanzi.
	MsgFmt uint8

	// SrcTerminalID is the source terminal MSISDN (32 bytes). For status reports, it's the destination terminal number from the Submit PDU.
	SrcTerminalID string `safe_json:"-"`

	// SrcTerminalType is the source terminal number type (1 byte): 0=Real number, 1=Pseudo code.
	SrcTerminalType uint8

	// RegisteredDeliver indicates if it's a status report (1 byte): 0=Not a status report (MO), 1=Status report (receipt).
	RegisteredDeliver uint8

	// MsgLength is the message length (1 byte).
	MsgLength uint8

	// MsgContent is the message content (MsgLength bytes).
	MsgContent string

	// LinkID is used for on-demand services (20 bytes). Not used in non-on-demand MT processes.
	LinkID string
}

// IDecode decodes the byte slice into a Deliver PDU.
func (d *Deliver) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return cmpp.ErrInvalidPudLength
	}

	b := packet.NewPacketReader(data)
	defer b.Release()

	d.Header = cmpp.ReadHeader(b)
	d.MsgID = b.ReadUint64()
	d.DestID = b.ReadCStringN(21)
	d.ServiceID = b.ReadCStringN(10)
	d.TpPID = b.ReadUint8()
	d.TpUDHI = b.ReadUint8()
	d.MsgFmt = b.ReadUint8()
	d.SrcTerminalID = b.ReadCStringN(32)
	d.SrcTerminalType = b.ReadUint8()
	d.RegisteredDeliver = b.ReadUint8()
	d.MsgLength = b.ReadUint8()
	d.MsgContent = string(b.ReadNBytes(int(d.MsgLength)))
	d.LinkID = b.ReadCStringN(20)

	return b.Error()
}

// IEncode encodes the Deliver PDU into a byte slice.
func (d *Deliver) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()

	cmpp.WriteHeaderNoLength(d.Header, b)
	b.WriteUint64(d.MsgID)
	b.WriteFixedLenString(d.DestID, 21)
	b.WriteFixedLenString(d.ServiceID, 10)
	b.WriteUint8(d.TpPID)
	b.WriteUint8(d.TpUDHI)
	b.WriteUint8(d.MsgFmt)
	b.WriteFixedLenString(d.SrcTerminalID, 32)
	b.WriteUint8(d.SrcTerminalType)
	b.WriteUint8(d.RegisteredDeliver)
	b.WriteUint8(d.MsgLength)
	b.WriteFixedLenString(d.MsgContent, int(d.MsgLength))
	b.WriteFixedLenString(d.LinkID, 20)

	return b.BytesWithLength()
}

// SetSequenceID sets the sequence ID of the PDU.
func (d *Deliver) SetSequenceID(id uint32) {
	d.Header.SequenceID = id
}

// GetSequenceID returns the sequence ID of the PDU.
func (d *Deliver) GetSequenceID() uint32 {
	return d.Header.SequenceID
}

// GetCommand returns the command ID of the PDU.
func (d *Deliver) GetCommand() sms.ICommander {
	return cmpp.CommandDeliver
}

// GenEmptyResponse generates an empty response PDU for the Deliver.
func (d *Deliver) GenEmptyResponse() sms.PDU {
	return &DeliverResp{
		Header: cmpp.Header{
			CommandID:  cmpp.CommandDeliverResp,
			SequenceID: d.GetSequenceID(),
		},
	}
}

// String returns a string representation of the Deliver PDU.
func (d *Deliver) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", d.Header)
	w.Write("MsgID", d.MsgID)
	w.Write("DestID", d.DestID)
	w.Write("ServiceID", d.ServiceID)
	w.Write("TpPID", d.TpPID)
	w.Write("TpUDHI", d.TpUDHI)
	w.Write("MsgFmt", d.MsgFmt)
	w.Write("SrcTerminalID", d.SrcTerminalID)
	w.Write("SrcTerminalType", d.SrcTerminalType)
	w.Write("RegisteredDeliver", d.RegisteredDeliver)
	w.Write("MsgLength", d.MsgLength)
	w.Write("MsgContent", d.MsgContent)
	w.Write("LinkID", d.LinkID)

	return w.String()
}

// -----------------------------------------------------------------------------------------------------

// DeliverResp represents a CMPP 3.0 DeliverResp PDU.
// It is the response to a Deliver PDU.
type DeliverResp struct {
	Header cmpp.Header
	// MsgID is the message identifier from the corresponding Deliver PDU (8 bytes).
	MsgID uint64

	// Result indicates the result of processing the Deliver PDU (4 bytes).
	// 0: Success
	// 1: Invalid message structure
	// 2: Invalid command ID
	// 3: Duplicate sequence ID
	// 4: Invalid message length
	// 5: Invalid fee code
	// 6: Message length exceeds maximum
	// 7: Invalid service ID
	// 8: Flow control error
	// 9+: Other errors
	Result uint32
}

// IDecode decodes the byte slice into a DeliverResp PDU.
func (d *DeliverResp) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return cmpp.ErrInvalidPudLength
	}
	b := packet.NewPacketReader(data)
	defer b.Release()

	d.Header = cmpp.ReadHeader(b)
	d.MsgID = b.ReadUint64()
	d.Result = b.ReadUint32()

	return b.Error()
}

// IEncode encodes the DeliverResp PDU into a byte slice.
func (d *DeliverResp) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()

	cmpp.WriteHeaderNoLength(d.Header, b)
	b.WriteUint64(d.MsgID)
	b.WriteUint32(d.Result)

	return b.BytesWithLength()
}

// SetSequenceID sets the sequence ID of the PDU.
func (d *DeliverResp) SetSequenceID(id uint32) {
	d.Header.SequenceID = id
}

// GetSequenceID returns the sequence ID of the PDU.
func (d *DeliverResp) GetSequenceID() uint32 {
	return d.Header.SequenceID
}

// GetCommand returns the command ID of the PDU.
func (d *DeliverResp) GetCommand() sms.ICommander {
	return cmpp.CommandDeliverResp
}

// GenEmptyResponse generates an empty response PDU (nil for DeliverResp).
func (d *DeliverResp) GenEmptyResponse() sms.PDU {
	return nil
}

// String returns a string representation of the DeliverResp PDU.
func (d *DeliverResp) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", d.Header)
	w.Write("MsgID", d.MsgID)
	w.Write("Result", d.Result)

	return w.String()
}
