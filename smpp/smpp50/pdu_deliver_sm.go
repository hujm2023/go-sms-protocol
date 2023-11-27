package smpp50

import (
	"github.com/hujm2023/go-sms-protocol/packet"
	"github.com/hujm2023/go-sms-protocol/smpp"
)

type DeliverSm struct {
	smpp.Header

	// CString, max 6
	ServiceType string

	SourceAddrTon uint8
	SourceAddrNpi uint8
	SourceAddr    string // CString, max 21

	DestAddrTon     uint8
	DestAddrNpi     uint8
	DestinationAddr string // CString, max 21

	ESMClass             uint8
	ProtocolID           uint8
	PriorityFlag         uint8
	ScheduleDeliveryTime string // ptime, 1 or max 17
	ValidityPeriod       string // ptime, 1 or max 17

	RegisteredDelivery   uint8
	ReplaceIfPresentFlag uint8
	DataCoding           uint8
	SmDefaultMsgID       uint8
	SmLength             uint8
	ShortMessage         []byte // max 255

	tlv smpp.TLVs
}

func (d *DeliverSm) IDecode(data []byte) error {
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
	d.DestinationAddr = buf.ReadCString()
	d.ESMClass = buf.ReadUint8()
	d.ProtocolID = buf.ReadUint8()
	d.PriorityFlag = buf.ReadUint8()
	d.ScheduleDeliveryTime = buf.ReadCString()
	d.ValidityPeriod = buf.ReadCString()
	d.RegisteredDelivery = buf.ReadUint8()
	d.ReplaceIfPresentFlag = buf.ReadUint8()
	d.DataCoding = buf.ReadUint8()
	d.SmDefaultMsgID = buf.ReadUint8()
	d.SmLength = buf.ReadUint8()
	d.ShortMessage = buf.ReadNBytes(int(d.SmLength))
	d.tlv = smpp.ReadTLVs1(buf)

	return buf.Error()
}

func (d *DeliverSm) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter()
	defer buf.Release()

	smpp.WriteHeaderNoLength(d.Header, buf)
	buf.WriteCString(d.ServiceType)
	buf.WriteUint8(d.SourceAddrTon)
	buf.WriteUint8(d.SourceAddrNpi)
	buf.WriteCString(d.SourceAddr)
	buf.WriteUint8(d.DestAddrTon)
	buf.WriteUint8(d.DestAddrNpi)
	buf.WriteCString(d.DestinationAddr)
	buf.WriteUint8(d.ESMClass)
	buf.WriteUint8(d.ProtocolID)
	buf.WriteUint8(d.PriorityFlag)
	buf.WriteCString(d.ScheduleDeliveryTime)
	buf.WriteCString(d.ValidityPeriod)
	buf.WriteUint8(d.RegisteredDelivery)
	buf.WriteUint8(d.ReplaceIfPresentFlag)
	buf.WriteUint8(d.DataCoding)
	buf.WriteUint8(d.SmDefaultMsgID)
	buf.WriteUint8(d.SmLength)
	buf.WriteBytes(d.ShortMessage)
	buf.WriteBytes(d.tlv.Bytes())

	return buf.BytesWithLength()
}

func (d *DeliverSm) SetSequenceID(id uint32) {
	d.Header.Sequence = id
}

type DeliverSmResp struct {
	smpp.Header

	// CString, max 65
	MessageID string

	tlvs smpp.TLVs
}

func (d *DeliverSmResp) IDecode(data []byte) error {
	if len(data) < smpp.MinSMPPPacketLen {
		return smpp.ErrInvalidPudLength
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	d.Header = smpp.ReadHeader(buf)
	d.MessageID = buf.ReadCString()
	d.tlvs = smpp.ReadTLVs1(buf)

	return buf.Error()
}

func (d *DeliverSmResp) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter()
	defer buf.Release()

	smpp.WriteHeaderNoLength(d.Header, buf)
	buf.WriteCString(d.MessageID)
	buf.WriteBytes(d.tlvs.Bytes())

	return buf.BytesWithLength()
}

func (d *DeliverSmResp) SetSequenceID(id uint32) {
	d.Header.Sequence = id
}
