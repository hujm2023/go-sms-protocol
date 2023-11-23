package smpp34

import (
	"github.com/hujm2023/go-sms-protocol/packet"
)

type Bind struct {
	Header

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
	if len(data) < MinSMPPPacketLen {
		return ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	b.Header = ReadHeader(buf)
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

	buf.WriteUint32(uint32(b.ID))
	buf.WriteUint32(uint32(b.Status))
	buf.WriteUint32(b.Sequence)
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
	Header

	// CString, max 16
	SystemID string

	tlv map[uint16]TLV
}

func (b *BindResp) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	buf.WriteUint32(uint32(b.ID))
	buf.WriteUint32(uint32(b.Status))
	buf.WriteUint32(b.Sequence)
	buf.WriteCString(b.SystemID)

	for _, tlv := range b.tlv {
		buf.WriteBytes(tlv.Bytes())
	}

	return buf.BytesWithLength()
}

func (b *BindResp) IDecode(data []byte) error {
	if len(data) < MinSMPPPacketLen {
		return ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	b.Header = ReadHeader(buf)
	b.SystemID = buf.ReadCString()
	tlv, err := ReadTLVs(buf)
	if err != nil {
		return err
	}
	b.tlv = tlv

	return buf.Error()
}

func (b *BindResp) SetSequenceID(id uint32) {
	b.Header.Sequence = id
}
