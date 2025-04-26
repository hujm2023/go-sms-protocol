package cmpp30

import (
	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/packet"
)

// Cancel represents a CMPP 3.0 Cancel PDU.
// It is used by the SP to request the deletion of a previously submitted short message.
type Cancel struct {
	cmpp.Header

	// MsgID is the message identifier of the message to be cancelled (8 bytes).
	MsgID uint64
}

// IDecode decodes the byte slice into a Cancel PDU.
func (c *Cancel) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return cmpp.ErrInvalidPudLength
	}

	b := packet.NewPacketReader(data)
	defer b.Release()

	c.Header = cmpp.ReadHeader(b)
	c.MsgID = b.ReadUint64()
	return b.Error()
}

// IEncode encodes the Cancel PDU into a byte slice.
func (c *Cancel) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()

	cmpp.WriteHeaderNoLength(c.Header, b)
	b.WriteUint64(c.MsgID)
	return b.BytesWithLength()
}

// SetSequenceID sets the sequence ID of the PDU.
func (c *Cancel) SetSequenceID(id uint32) {
	c.Header.SequenceID = id
}

// GetSequenceID returns the sequence ID of the PDU.
func (c *Cancel) GetSequenceID() uint32 {
	return c.Header.SequenceID
}

// GetCommand returns the command ID of the PDU.
func (c *Cancel) GetCommand() sms.ICommander {
	return cmpp.CommandCancel
}

// GenEmptyResponse generates an empty response PDU for the Cancel.
func (c *Cancel) GenEmptyResponse() sms.PDU {
	return &CancelResp{
		Header: cmpp.Header{
			CommandID:  cmpp.CommandCancelResp,
			SequenceID: c.GetSequenceID(),
		},
	}
}

// String returns a string representation of the Cancel PDU.
func (c *Cancel) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", c.Header)
	w.Write("MsgID", c.MsgID)

	return w.String()
}

// CancelResp represents a CMPP 3.0 CancelResp PDU.
// It is the response to a Cancel PDU.
type CancelResp struct {
	Header cmpp.Header

	// SuccessID indicates the result of the cancel operation (4 bytes): 0 for success, 1 for failure.
	SuccessID uint32
}

// IDecode decodes the byte slice into a CancelResp PDU.
func (c *CancelResp) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return cmpp.ErrInvalidPudLength
	}
	b := packet.NewPacketReader(data)
	defer b.Release()

	c.Header = cmpp.ReadHeader(b)
	c.SuccessID = b.ReadUint32()
	return b.Error()
}

// IEncode encodes the CancelResp PDU into a byte slice.
func (c *CancelResp) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()

	cmpp.WriteHeaderNoLength(c.Header, b)
	b.WriteUint32(c.SuccessID)
	return b.BytesWithLength()
}

// SetSequenceID sets the sequence ID of the PDU.
func (c *CancelResp) SetSequenceID(id uint32) {
	c.Header.SequenceID = id
}

// GetSequenceID returns the sequence ID of the PDU.
func (c *CancelResp) GetSequenceID() uint32 {
	return c.Header.SequenceID
}

// GetCommand returns the command ID of the PDU.
func (c *CancelResp) GetCommand() sms.ICommander {
	return cmpp.CommandCancelResp
}

// GenEmptyResponse generates an empty response PDU (nil for CancelResp).
func (c *CancelResp) GenEmptyResponse() sms.PDU {
	return nil
}

// String returns a string representation of the CancelResp PDU.

func (c *CancelResp) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", c.Header)
	w.Write("SuccessID", c.SuccessID)

	return w.String()
}
