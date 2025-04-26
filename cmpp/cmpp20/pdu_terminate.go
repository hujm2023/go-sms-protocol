package cmpp20

import (
	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/packet"
)

// PduTerminate represents a CMPP Terminate PDU.
// It is used to terminate a connection.
type PduTerminate struct {
	cmpp.Header
}

// IEncode encodes the PduTerminate PDU into a byte slice.
func (p *PduTerminate) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	// header
	cmpp.WriteHeaderNoLength(p.Header, buf)

	return buf.BytesWithLength()
}

// IDecode decodes the byte slice into a PduTerminate PDU.
func (p *PduTerminate) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	p.Header = cmpp.ReadHeader(buf)

	return buf.Error()
}

// GetSequenceID returns the sequence ID of the PDU.
func (p *PduTerminate) GetSequenceID() uint32 {
	return p.Header.SequenceID
}

// SetSequenceID sets the sequence ID of the PDU.
func (p *PduTerminate) SetSequenceID(id uint32) {
	p.Header.SequenceID = id
}

// GetCommand returns the command ID of the PDU.
func (p *PduTerminate) GetCommand() sms.ICommander {
	return cmpp.CommandTerminate
}

// GenEmptyResponse generates an empty response PDU for the PduTerminate.
func (p *PduTerminate) GenEmptyResponse() sms.PDU {
	return &PduTerminateResp{
		Header: cmpp.Header{
			CommandID:  cmpp.CommandTerminateResp,
			SequenceID: p.GetSequenceID(),
		},
	}
}

// String returns a string representation of the PduTerminate PDU.
func (p *PduTerminate) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", p.Header)

	return w.String()
}

// --------------

// PduTerminateResp represents a CMPP TerminateResp PDU.
// It is the response to a PduTerminate.
type PduTerminateResp struct {
	cmpp.Header
}

// IEncode encodes the PduTerminateResp PDU into a byte slice.
func (p *PduTerminateResp) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	// header
	cmpp.WriteHeaderNoLength(p.Header, buf)

	return buf.Bytes()
}

// IDecode decodes the byte slice into a PduTerminateResp PDU.
func (p *PduTerminateResp) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	p.Header = cmpp.ReadHeader(buf)

	return buf.Error()
}

// GetSequenceID returns the sequence ID of the PDU.
func (p *PduTerminateResp) GetSequenceID() uint32 {
	return p.Header.SequenceID
}

// SetSequenceID sets the sequence ID of the PDU.
func (p *PduTerminateResp) SetSequenceID(id uint32) {
	p.Header.SequenceID = id
}

// GetCommand returns the command ID of the PDU.
func (p *PduTerminateResp) GetCommand() sms.ICommander {
	return cmpp.CommandTerminateResp
}

// GenEmptyResponse generates an empty response PDU (nil for TerminateResp).
func (p *PduTerminateResp) GenEmptyResponse() sms.PDU {
	return nil
}

// String returns a string representation of the PduTerminateResp PDU.
func (p *PduTerminateResp) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", p.Header)

	return w.String()
}
