package smpp34

import (
	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/packet"
	"github.com/hujm2023/go-sms-protocol/smpp"
)

type Bind struct {
	smpp.Header

	// CString, max 16,Identifies the ESME system requesting to bind as a receiver with the SMSC.
	SystemID string

	// CString, max 9
	Password string

	// CString, max 13
	SystemType string

	InterfaceVersion uint8

	AddrTon uint8

	AddrNpi uint8

	// CString, max 41
	AddressRange string
}

func (b *Bind) GetSequenceID() uint32 {
	return b.Header.Sequence
}

func (b *Bind) GetCommand() sms.ICommander {
	return smpp.BIND_TRANSCEIVER
}

func (b *Bind) GenEmptyResponse() sms.PDU {
	return &BindResp{
		Header: smpp.Header{
			ID:       smpp.BIND_TRANSCEIVER_RESP,
			Sequence: b.Header.Sequence,
		},
	}
}

func (b *Bind) IDecode(data []byte) error {
	if len(data) < smpp.MinSMPPPacketLen {
		return smpp.ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	b.Header = smpp.ReadHeader(buf)
	b.SystemID = buf.ReadCString()
	b.Password = buf.ReadCString()
	b.SystemType = buf.ReadCString()
	b.InterfaceVersion = buf.ReadUint8()
	b.AddrTon = buf.ReadUint8()
	b.AddrNpi = buf.ReadUint8()
	b.AddressRange = buf.ReadCString()

	return buf.Error()
}

func (b *Bind) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	smpp.WriteHeaderNoLength(b.Header, buf)

	buf.WriteCString(b.SystemID)
	buf.WriteCString(b.Password)
	buf.WriteCString(b.SystemType)
	buf.WriteUint8(b.InterfaceVersion)
	buf.WriteUint8(b.AddrTon)
	buf.WriteUint8(b.AddrNpi)
	buf.WriteCString(b.AddressRange)

	return buf.BytesWithLength()
}

func (b *Bind) SetSequenceID(id uint32) {
	b.Header.Sequence = id
}

func (b *Bind) String() string {
	s := packet.NewPDUStringer()
	defer s.Release()

	s.Write("Header", b.Header)
	s.Write("SystemID", b.SystemID)
	s.Write("Password", b.Password)
	s.Write("SystemType", b.SystemType)
	s.Write("InterfaceVersion", b.InterfaceVersion)
	s.Write("AddrTon", b.AddrTon)
	s.Write("AddrNpi", b.AddrNpi)
	s.Write("AddressRange", b.AddressRange)

	return s.String()
}

type BindResp struct {
	smpp.Header

	// CString, max 16
	SystemID string

	TLVs smpp.TLVs
}

func (b *BindResp) GetSequenceID() uint32 {
	return b.Header.Sequence
}

func (b *BindResp) GetCommand() sms.ICommander {
	return smpp.BIND_TRANSCEIVER_RESP
}

func (b *BindResp) GenEmptyResponse() sms.PDU {
	return nil
}

func (b *BindResp) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	smpp.WriteHeaderNoLength(b.Header, buf)

	buf.WriteCString(b.SystemID)

	buf.WriteBytes(b.TLVs.Bytes())

	return buf.BytesWithLength()
}

func (b *BindResp) IDecode(data []byte) error {
	if len(data) < smpp.MinSMPPPacketLen {
		return smpp.ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	b.Header = smpp.ReadHeader(buf)

	b.SystemID = buf.ReadCString()

	b.TLVs = smpp.ReadTLVs1(buf)

	return buf.Error()
}

func (b *BindResp) SetSequenceID(id uint32) {
	b.Header.Sequence = id
}

func (b *BindResp) String() string {
	s := packet.NewPDUStringer()
	defer s.Release()

	s.Write("Header", b.Header)
	s.Write("SystemID", b.SystemID)
	s.OmitWrite("TLVs", b.TLVs.String())

	return s.String()
}
