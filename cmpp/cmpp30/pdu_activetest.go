package cmpp30

import (
	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/packet"
)

// ActiveTest represents a CMPP 3.0 ActiveTest PDU.
// It is used to test the connection status between the SP and the ISMG.
type ActiveTest struct {
	cmpp.Header
}

// IEncode encodes the ActiveTest PDU into a byte slice.
func (p *ActiveTest) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter()
	defer buf.Release()

	cmpp.WriteHeaderNoLength(p.Header, buf)

	return buf.BytesWithLength()
}

// IDecode decodes the byte slice into an ActiveTest PDU.
func (p *ActiveTest) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return cmpp.ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	p.Header = cmpp.ReadHeader(buf)

	return buf.Error()
}

// SetSequenceID sets the sequence ID of the PDU.
func (p *ActiveTest) SetSequenceID(id uint32) {
	p.Header.SequenceID = id
}

// GetSequenceID returns the sequence ID of the PDU.
func (a *ActiveTest) GetSequenceID() uint32 {
	return a.Header.SequenceID
}

// GetCommand returns the command ID of the PDU.
func (a *ActiveTest) GetCommand() sms.ICommander {
	return cmpp.CommandActiveTest
}

// GenEmptyResponse generates an empty response PDU for the ActiveTest.
func (a *ActiveTest) GenEmptyResponse() sms.PDU {
	return &ActiveTestResp{
		Header: cmpp.Header{
			CommandID:  cmpp.CommandActiveTestResp,
			SequenceID: a.GetSequenceID(),
		},
	}
}

// String returns a string representation of the ActiveTest PDU.
func (a *ActiveTest) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", a.Header)

	return w.String()
}

// ActiveTestResp represents a CMPP 3.0 ActiveTestResp PDU.
// It is the response to an ActiveTest PDU.
type ActiveTestResp struct {
	cmpp.Header

	// Reserved is a reserved field (1 byte).
	Reserved uint8
}

// IEncode encodes the ActiveTestResp PDU into a byte slice.
func (pr *ActiveTestResp) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter()
	defer buf.Release()

	cmpp.WriteHeaderNoLength(pr.Header, buf)
	buf.WriteUint8(pr.Reserved)

	return buf.BytesWithLength()
}

// IDecode decodes the byte slice into an ActiveTestResp PDU.
func (pr *ActiveTestResp) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return cmpp.ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	pr.Header = cmpp.ReadHeader(buf)
	pr.Reserved = buf.ReadUint8()

	return buf.Error()
}

// SetSequenceID sets the sequence ID of the PDU.
func (pr *ActiveTestResp) SetSequenceID(sid uint32) {
	pr.Header.SequenceID = sid
}

// GetSequenceID returns the sequence ID of the PDU.
func (a *ActiveTestResp) GetSequenceID() uint32 {
	return a.Header.SequenceID
}

// GetCommand returns the command ID of the PDU.
func (a *ActiveTestResp) GetCommand() sms.ICommander {
	return cmpp.CommandActiveTestResp
}

// GenEmptyResponse generates an empty response PDU (nil for ActiveTestResp).
func (a *ActiveTestResp) GenEmptyResponse() sms.PDU {
	return nil
}

// String returns a string representation of the ActiveTestResp PDU.
func (a *ActiveTestResp) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", a.Header)
	w.Write("Reserved", a.Reserved)

	return w.String()
}
