package cmpp30

import (
	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/packet"
)

// Terminate represents a CMPP 3.0 Terminate PDU.
// It is used by the SP or ISMG to terminate a connection.
type Terminate struct {
	cmpp.Header
}

// IDecode decodes the byte slice into a Terminate PDU.
func (t *Terminate) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return cmpp.ErrInvalidPudLength
	}
	b := packet.NewPacketReader(data)
	defer b.Release()

	t.Header = cmpp.ReadHeader(b)
	return b.Error()
}

// IEncode encodes the Terminate PDU into a byte slice.
func (t *Terminate) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()

	cmpp.WriteHeaderNoLength(t.Header, b)

	return b.BytesWithLength()
}

// SetSequenceID sets the sequence ID of the PDU.
func (t *Terminate) SetSequenceID(id uint32) {
	t.Header.SequenceID = id
}

// GetSequenceID returns the sequence ID of the PDU.
func (t *Terminate) GetSequenceID() uint32 {
	return t.Header.SequenceID
}

// GetCommand returns the command ID of the PDU.
func (t *Terminate) GetCommand() sms.ICommander {
	return cmpp.CommandTerminate
}

// GenEmptyResponse generates an empty response PDU for the Terminate.
func (t *Terminate) GenEmptyResponse() sms.PDU {
	return &TerminateResp{
		Header: cmpp.Header{
			CommandID:  cmpp.CommandTerminateResp,
			SequenceID: t.GetSequenceID(),
		},
	}
}

// String returns a string representation of the Terminate PDU.
func (t *Terminate) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", t.Header)

	return w.String()
}

// TerminateResp represents a CMPP 3.0 TerminateResp PDU.
// It is the response to a Terminate PDU.
type TerminateResp struct {
	Header cmpp.Header
}

// IDecode decodes the byte slice into a TerminateResp PDU.
func (t *TerminateResp) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return cmpp.ErrInvalidPudLength
	}
	b := packet.NewPacketReader(data)
	defer b.Release()

	t.Header = cmpp.ReadHeader(b)
	return b.Error()
}

// IEncode encodes the TerminateResp PDU into a byte slice.
func (t *TerminateResp) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()

	cmpp.WriteHeaderNoLength(t.Header, b)

	return b.BytesWithLength()
}

// SetSequenceID sets the sequence ID of the PDU.
func (t *TerminateResp) SetSequenceID(id uint32) {
	t.Header.SequenceID = id
}

// GetSequenceID returns the sequence ID of the PDU.
func (t *TerminateResp) GetSequenceID() uint32 {
	return t.Header.SequenceID
}

// GetCommand returns the command ID of the PDU.
func (t *TerminateResp) GetCommand() sms.ICommander {
	return cmpp.CommandTerminateResp
}

// GenEmptyResponse generates an empty response PDU for the TerminateResp.
// TerminateResp does not have a response, so it returns nil.
func (t *TerminateResp) GenEmptyResponse() sms.PDU {
	return nil
}

// String returns a string representation of the TerminateResp PDU.
func (t *TerminateResp) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", t.Header)

	return w.String()
}
