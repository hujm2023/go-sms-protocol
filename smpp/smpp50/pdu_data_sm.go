package smpp50

import (
	"github.com/hujm2023/go-sms-protocol/packet"
	"github.com/hujm2023/go-sms-protocol/smpp"
)

type DataSM struct {
	smpp.Header

	ServiceType string // CString, max 6

	SourceAddrTon uint8
	SourceAddrNpi uint8
	SourceAddr    string // CString, max 65

	DestAddrTon     uint8
	DestAddrNpi     uint8
	DestenationAddr string // CString, max 65

	ESMClass           uint8
	RegisteredDelivery uint8
	DataCoding         uint8

	tlvs smpp.TLVs
}

func (d *DataSM) IDecode(data []byte) error {
	if len(data) < smpp.MinSMPPPacketLen {
		return smpp.ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	d.Header = smpp.ReadHeader(buf)
	d.ServiceType = buf.ReadCString()
	d.SourceAddrTon = buf.ReadUint8()
	d.SourceAddrNpi = buf.ReadUint8()
	d.SourceAddr = buf.ReadCString()
	d.DestAddrTon = buf.ReadUint8()
	d.DestAddrNpi = buf.ReadUint8()
	d.DestenationAddr = buf.ReadCString()
	d.ESMClass = buf.ReadUint8()
	d.RegisteredDelivery = buf.ReadUint8()
	d.DataCoding = buf.ReadUint8()
	d.tlvs = smpp.ReadTLVs1(buf)

	return buf.Error()
}

func (d *DataSM) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter()
	defer buf.Release()

	smpp.WriteHeaderNoLength(d.Header, buf)
	buf.WriteCString(d.ServiceType)
	buf.WriteUint8(d.SourceAddrTon)
	buf.WriteUint8(d.SourceAddrNpi)
	buf.WriteCString(d.SourceAddr)
	buf.WriteUint8(d.DestAddrTon)
	buf.WriteUint8(d.DestAddrNpi)
	buf.WriteCString(d.DestenationAddr)
	buf.WriteUint8(d.ESMClass)
	buf.WriteUint8(d.RegisteredDelivery)
	buf.WriteUint8(d.DataCoding)
	buf.WriteBytes(d.tlvs.Bytes())

	return buf.BytesWithLength()
}

func (d *DataSM) SetSequenceID(id uint32) {
	d.Header.Sequence = id
}
