package smpp34

import (
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

type BindResp struct {
	smpp.Header

	// CString, max 16
	SystemID string

	tlv smpp.TLVs
}

func (b *BindResp) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	smpp.WriteHeaderNoLength(b.Header, buf)

	buf.WriteCString(b.SystemID)

	buf.WriteBytes(b.tlv.Bytes())

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

	b.tlv = smpp.ReadTLVs1(buf)

	return buf.Error()
}

func (b *BindResp) SetSequenceID(id uint32) {
	b.Header.Sequence = id
}
