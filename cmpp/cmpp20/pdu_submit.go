package cmpp20

import (
	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/packet"
)

// PduSubmit represents a CMPP 2.0 Submit PDU.
// It is used by the SP to submit a short message to the ISMG.
type PduSubmit struct {
	cmpp.Header

	// MsgID is the message identifier (8 bytes), generated by the SP gateway, left blank here.
	MsgID uint64

	// PkTotal is the total number of packets for the same MsgID (1 byte, starts from 1).
	PkTotal uint8

	// PkNumber is the sequence number for the same MsgID (1 byte, starts from 1).
	PkNumber uint8

	// RegisteredDelivery indicates if a status report is required (1 byte): 0=No, 1=Yes, 2=Generate SMC bill (for billing only, not sent to terminal).
	RegisteredDelivery uint8

	// MsgLevel is the message priority level (1 byte).
	MsgLevel uint8

	// ServiceID is the service type (10 bytes), a combination of digits, letters, and symbols.
	ServiceID string

	// FeeUserType indicates the billing user type (1 byte): 0=Destination terminal, 1=Source terminal, 2=SP, 3=Field invalid (refer to FeeTerminalID).
	FeeUserType uint8

	// FeeTerminalID is the billed user's number (21 bytes). If blank, field is invalid (refer to FeeUserType, mutually exclusive).
	FeeTerminalID string

	// TpPID is the GSM protocol type (1 byte). See GSM 03.40 section 9.2.3.9.
	TpPID uint8

	// TpUDHI is the GSM protocol type (1 byte). See GSM 03.40 section 9.2.3.23 (only 1 bit used, right-aligned).
	TpUDHI uint8

	// MsgFmt is the message format (1 byte): 0=ASCII, 3=SMS Write Card, 4=Binary, 8=UCS2, 15=GB Hanzi.
	MsgFmt uint8

	// MsgSrc is the message source (SP ID) (6 bytes).
	MsgSrc string

	// FeeType is the fee category (2 bytes):
	// 01: Free for 'FeeTerminalID'
	// 02: Per-message fee for 'FeeTerminalID'
	// 03: Monthly fee for 'FeeTerminalID'
	// 04: Capped fee for 'FeeTerminalID'
	// 05: Fee handled by SP for 'FeeTerminalID'
	FeeType string

	// FeeCode is the fee code (6 bytes, in cents).
	FeeCode string

	// ValIDTime is the validity period (17 bytes, SMPP 3.3 format).
	ValIDTime string

	// AtTime is the scheduled delivery time (17 bytes, SMPP 3.3 format).
	AtTime string

	// SrcID is the source number (21 bytes). SP's service code or long number prefixed with it. Displayed as sender on user's phone.
	// Corresponds to userExt field in our implementation.
	SrcID string

	// DestUsrTL is the number of recipient users (1 byte, < 100).
	DestUsrTL uint8

	// DestTerminalID is the list of recipient MSISDNs (21 * DestUsrTL bytes). Each MSISDN is 21 bytes.
	DestTerminalID []string

	// MsgLength is the message length (1 byte). MsgFmt=0: <160 bytes; Others: <=140 bytes.
	MsgLength uint8

	// MsgContent is the message content (MsgLength bytes).
	MsgContent []byte

	// Reserve is a reserved field (8 bytes).
	Reserve string
}

// IEncode encodes the PduSubmit PDU into a byte slice.
func (p *PduSubmit) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter(0)
	defer b.Release()

	cmpp.WriteHeaderNoLength(p.Header, b)
	b.WriteUint64(p.MsgID)
	b.WriteUint8(p.PkTotal)
	b.WriteUint8(p.PkNumber)
	b.WriteUint8(p.RegisteredDelivery)
	b.WriteUint8(p.MsgLevel)
	b.WriteFixedLenString(p.ServiceID, 10)
	b.WriteUint8(p.FeeUserType)
	b.WriteFixedLenString(p.FeeTerminalID, 21)
	b.WriteUint8(p.TpPID)
	b.WriteUint8(p.TpUDHI)
	b.WriteUint8(p.MsgFmt)
	b.WriteFixedLenString(p.MsgSrc, 6)
	b.WriteFixedLenString(p.FeeType, 2)
	b.WriteFixedLenString(p.FeeCode, 6)
	b.WriteFixedLenString(p.ValIDTime, 17)
	b.WriteFixedLenString(p.AtTime, 17)
	b.WriteFixedLenString(p.SrcID, 21)
	b.WriteUint8(p.DestUsrTL)
	for _, dest := range p.DestTerminalID {
		b.WriteFixedLenString(dest, 21)
	}
	b.WriteUint8(p.MsgLength)
	b.WriteBytes(p.MsgContent)
	b.WriteFixedLenString(p.Reserve, 8)

	return b.BytesWithLength()
}

// IDecode decodes the byte slice into a PduSubmit PDU.
func (p *PduSubmit) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return ErrInvalidPudLength
	}
	b := packet.NewPacketReader(data)
	defer b.Release()

	p.Header = cmpp.ReadHeader(b)
	p.MsgID = b.ReadUint64()
	p.PkTotal = b.ReadUint8()
	p.PkNumber = b.ReadUint8()
	p.RegisteredDelivery = b.ReadUint8()
	p.MsgLevel = b.ReadUint8()
	p.ServiceID = b.ReadCStringN(10)
	p.FeeUserType = b.ReadUint8()
	p.FeeTerminalID = b.ReadCStringN(21)
	p.TpPID = b.ReadUint8()
	p.TpUDHI = b.ReadUint8()
	p.MsgFmt = b.ReadUint8()
	p.MsgSrc = b.ReadCStringN(6)
	p.FeeType = b.ReadCStringN(2)
	p.FeeCode = b.ReadCStringN(6)
	p.ValIDTime = b.ReadCStringN(17)
	p.AtTime = b.ReadCStringN(17)
	p.SrcID = b.ReadCStringN(21)
	p.DestUsrTL = b.ReadUint8()
	p.DestTerminalID = make([]string, p.DestUsrTL)
	for i := 0; i < int(p.DestUsrTL); i++ {
		p.DestTerminalID[i] = b.ReadCStringN(21)
	}
	p.MsgLength = b.ReadUint8()
	p.MsgContent = b.ReadNBytes(int(p.MsgLength))
	p.Reserve = b.ReadCStringN(8)

	return b.Error()
}

// GetSequenceID returns the sequence ID of the PDU.
func (p *PduSubmit) GetSequenceID() uint32 {
	return p.Header.SequenceID
}

// SetSequenceID sets the sequence ID of the PDU.
func (p *PduSubmit) SetSequenceID(id uint32) {
	p.Header.SequenceID = id
}

// GetCommand returns the command ID of the PDU.
func (p *PduSubmit) GetCommand() sms.ICommander {
	return cmpp.CommandSubmit
}

// GenEmptyResponse generates an empty response PDU for the PduSubmit.
func (p *PduSubmit) GenEmptyResponse() sms.PDU {
	return &PduSubmitResp{
		Header: cmpp.Header{
			CommandID:  cmpp.CommandSubmitResp,
			SequenceID: p.GetSequenceID(),
		},
	}
}

// String returns a string representation of the PduSubmit PDU.
func (p *PduSubmit) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", p.Header)
	w.Write("MsgID", p.MsgID)
	w.Write("PkTotal", p.PkTotal)
	w.Write("PkNumber", p.PkNumber)
	w.Write("RegisteredDelivery", p.RegisteredDelivery)
	w.Write("MsgLevel", p.MsgLevel)
	w.Write("ServiceID", p.ServiceID)
	w.Write("FeeUserType", p.FeeUserType)
	w.Write("FeeTerminalID", p.FeeTerminalID)
	w.Write("TpPID", p.TpPID)
	w.Write("TpUDHI", p.TpUDHI)
	w.Write("MsgFmt", p.MsgFmt)
	w.Write("MsgSrc", p.MsgSrc)
	w.Write("FeeType", p.FeeType)
	w.Write("FeeCode", p.FeeCode)
	w.Write("ValIDTime", p.ValIDTime)
	w.Write("AtTime", p.AtTime)
	w.Write("SrcID", p.SrcID)
	w.Write("DestUsrTL", p.DestUsrTL)
	w.Write("DestTerminalID", p.DestTerminalID)
	w.Write("MsgLength", p.MsgLength)
	w.WriteWithBytes("MsgContent", p.MsgContent)
	w.Write("Reserve", p.Reserve)

	return w.String()
}

// PduSubmitResp represents a CMPP 2.0 SubmitResp PDU.
// It is the response to a Submit PDU.
type PduSubmitResp struct {
	Header cmpp.Header
	// MsgID is the message identifier assigned by the ISMG (8 bytes).
	MsgID uint64

	// Result indicates the result of processing the Submit PDU (1 byte).
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
	Result uint8
}

// IEncode encodes the PduSubmitResp PDU into a byte slice.
func (pr *PduSubmitResp) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter(0)
	defer b.Release()

	cmpp.WriteHeaderNoLength(pr.Header, b)
	b.WriteUint64(pr.MsgID)
	b.WriteUint8(pr.Result)

	return b.BytesWithLength()
}

// IDecode decodes the byte slice into a PduSubmitResp PDU.
func (pr *PduSubmitResp) IDecode(data []byte) error {
	b := packet.NewPacketReader(data)
	defer b.Release()

	pr.Header = cmpp.ReadHeader(b)
	pr.MsgID = b.ReadUint64()
	pr.Result = b.ReadUint8()

	return b.Error()
}

// GetSequenceID returns the sequence ID of the PDU.
func (pr *PduSubmitResp) GetSequenceID() uint32 {
	return pr.Header.SequenceID
}

// SetSequenceID sets the sequence ID of the PDU.
func (pr *PduSubmitResp) SetSequenceID(id uint32) {
	pr.Header.SequenceID = id
}

// GetCommand returns the command ID of the PDU.
func (p *PduSubmitResp) GetCommand() sms.ICommander {
	return cmpp.CommandSubmitResp
}

// GenEmptyResponse generates an empty response PDU (nil for PduSubmitResp).
func (p *PduSubmitResp) GenEmptyResponse() sms.PDU {
	return nil
}

// String returns a string representation of the PduSubmitResp PDU.
func (p *PduSubmitResp) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", p.Header)
	w.Write("MsgID", p.MsgID)
	w.Write("Result", p.Result)

	return w.String()
}

var submitResultString = map[uint8]string{
	0: "Success",
	1: "Invalid message structure",
	2: "Invalid command ID",
	3: "Duplicate sequence ID",
	4: "Invalid message length",
	5: "Invalid fee code",
	6: "Message length exceeds maximum",
	7: "Invalid service ID",
	8: "Flow control error",
	9: "Other error",
}

// SubmitRespResultString returns the string representation of a SubmitResp result code.
func SubmitRespResultString(n uint8) string {
	v, ok := submitResultString[n]
	if ok {
		return v
	}
	return "Other error"
}
