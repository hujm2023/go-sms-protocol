package cmpp20

import (
	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/packet"
)

// PduActiveTest represents a CMPP 2.0 ActiveTest PDU.
// It is used to test the connection status between the SP and the ISMG.
type PduActiveTest struct {
	cmpp.Header
}

// IEncode encodes the PduActiveTest PDU into a byte slice.
func (p *PduActiveTest) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	cmpp.WriteHeaderNoLength(p.Header, buf)

	return buf.BytesWithLength()
}

// IDecode decodes the byte slice into a PduActiveTest PDU.
func (p *PduActiveTest) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	p.Header = cmpp.ReadHeader(buf)

	return buf.Error()
}

// GetSequenceID returns the sequence ID of the PDU.
func (p *PduActiveTest) GetSequenceID() uint32 {
	return p.Header.SequenceID
}

// SetSequenceID sets the sequence ID of the PDU.
func (p *PduActiveTest) SetSequenceID(sid uint32) {
	p.Header.SequenceID = sid
}

// GetCommand returns the command ID of the PDU.
func (p *PduActiveTest) GetCommand() sms.ICommander {
	return cmpp.CommandActiveTest
}

// GenEmptyResponse generates an empty response PDU for the PduActiveTest.
func (p *PduActiveTest) GenEmptyResponse() sms.PDU {
	return &PduActiveTestResp{
		Header: cmpp.Header{
			CommandID:  cmpp.CommandActiveTestResp,
			SequenceID: p.GetSequenceID(),
		},
	}
}

// String returns a string representation of the PduActiveTest PDU.
func (p *PduActiveTest) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", p.Header)

	return w.String()
}

// --------------------------------------------------------------------

// PduActiveTestResp represents a CMPP 2.0 ActiveTestResp PDU.
// It is the response to a PduActiveTest.
type PduActiveTestResp struct {
	cmpp.Header

	// Reserved is a reserved field (1 byte).
	Reserved uint8
}

// IEncode encodes the PduActiveTestResp PDU into a byte slice.
func (pr *PduActiveTestResp) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	cmpp.WriteHeaderNoLength(pr.Header, buf)
	buf.WriteUint8(pr.Reserved)

	return buf.BytesWithLength()
}

// IDecode decodes the byte slice into a PduActiveTestResp PDU.
func (pr *PduActiveTestResp) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	pr.Header = cmpp.ReadHeader(buf)
	pr.Reserved = buf.ReadUint8()

	return buf.Error()
}

// GetSequenceID returns the sequence ID of the PDU.
func (pr *PduActiveTestResp) GetSequenceID() uint32 {
	return pr.Header.SequenceID
}

// SetSequenceID sets the sequence ID of the PDU.
func (pr *PduActiveTestResp) SetSequenceID(sid uint32) {
	pr.Header.SequenceID = sid
}

// GetCommand returns the command ID of the PDU.
func (p *PduActiveTestResp) GetCommand() sms.ICommander {
	return cmpp.CommandActiveTestResp
}

// GenEmptyResponse generates an empty response PDU (nil for ActiveTestResp).
func (p *PduActiveTestResp) GenEmptyResponse() sms.PDU {
	return nil
}

// String returns a string representation of the PduActiveTestResp PDU.
func (p *PduActiveTestResp) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", p.Header)
	w.Write("Reserved", p.Reserved)

	return w.String()
}
