package smpp34

import (
	"github.com/hujm2023/go-sms-protocol/packet"
)

type DeliverSm struct {
	Header

	// CString, max 6
	ServiceType string

	SourceAddrTon uint8
	SourceAddrNpi uint8
	// CString, max 21
	SourceAddr string

	DestinationAddrTon uint8
	DestinationAddrNpi uint8
	// CString, max 21
	DestinationAddr string

	ESMClass     uint8
	ProtocolID   uint8
	PriorityFlag uint8

	// CString, max 17
	ScheduleDeliveryTime string

	// CString, max 17
	ValidityPeriod string

	RegisteredDelivery   uint8
	ReplaceIfPresentFlag uint8
	DataCoding           uint8
	SMDefaultMsgId       uint8

	SmLength     uint8
	ShortMessage []byte

	tlvs map[uint16]TLV
}

func (d *DeliverSm) IDecode(data []byte) error {
	if len(data) < MinSMPPPacketLen {
		return ErrInvalidPudLength
	}
	buf := packet.NewPacketReader(data)
	defer buf.Release()

	d.Header = ReadHeader(buf)
	d.ServiceType = buf.ReadCString()
	d.SourceAddrTon = buf.ReadUint8()
	d.SourceAddrNpi = buf.ReadUint8()
	d.SourceAddr = buf.ReadCString()
	d.DestinationAddrTon = buf.ReadUint8()
	d.DestinationAddrNpi = buf.ReadUint8()
	d.DestinationAddr = buf.ReadCString()
	d.ESMClass = buf.ReadUint8()
	d.ProtocolID = buf.ReadUint8()
	d.PriorityFlag = buf.ReadUint8()
	d.ScheduleDeliveryTime = buf.ReadCString()
	d.ValidityPeriod = buf.ReadCString()
	d.RegisteredDelivery = buf.ReadUint8()
	d.ReplaceIfPresentFlag = buf.ReadUint8()
	d.DataCoding = buf.ReadUint8()
	d.SMDefaultMsgId = buf.ReadUint8()
	d.SmLength = buf.ReadUint8()
	temp := make([]byte, d.SmLength)
	buf.ReadBytes(temp)
	d.ShortMessage = temp

	tlv, err := ReadTLVs(buf)
	if err != nil {
		return err
	}
	d.tlvs = tlv
	return buf.Error()
}

func (d *DeliverSm) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	buf.WriteUint32(uint32(d.Header.ID))
	buf.WriteUint32(uint32(d.Header.Status))
	buf.WriteUint32(d.Header.Sequence)

	buf.WriteCString(d.ServiceType)
	buf.WriteUint8(d.SourceAddrTon)
	buf.WriteUint8(d.SourceAddrNpi)
	buf.WriteCString(d.SourceAddr)
	buf.WriteUint8(d.DestinationAddrTon)
	buf.WriteUint8(d.DestinationAddrNpi)
	buf.WriteCString(d.DestinationAddr)
	buf.WriteUint8(d.ESMClass)
	buf.WriteUint8(d.ProtocolID)
	buf.WriteUint8(d.PriorityFlag)
	buf.WriteCString(d.ScheduleDeliveryTime)
	buf.WriteCString(d.ValidityPeriod)
	buf.WriteUint8(d.RegisteredDelivery)
	buf.WriteUint8(d.ReplaceIfPresentFlag)
	buf.WriteUint8(d.DataCoding)
	buf.WriteUint8(d.SMDefaultMsgId)
	buf.WriteUint8(d.SmLength)
	buf.WriteBytes(d.ShortMessage)

	for _, tlv := range d.tlvs {
		buf.WriteBytes(tlv.Bytes())
	}

	return buf.BytesWithLength()
}

func (d *DeliverSm) SetSequenceID(id uint32) {
	d.Header.Sequence = id
}

type DeliverSMResp struct {
	Header

	// CString, size 1, unused, set to null
	MessageID string
}

func (d *DeliverSMResp) IDecode(data []byte) error {
	if len(data) < MinSMPPPacketLen {
		return ErrInvalidPudLength
	}
	buf := packet.NewPacketReader(data)
	defer buf.Release()

	d.Header = ReadHeader(buf)
	d.MessageID = buf.ReadCString()
	return buf.Error()
}

func (d *DeliverSMResp) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	buf.WriteUint32(uint32(d.Header.ID))
	buf.WriteUint8(uint8(d.Header.Status))
	buf.WriteUint32(d.Header.Sequence)

	buf.WriteCString(d.MessageID)

	return buf.BytesWithLength()
}

func (d *DeliverSMResp) SetSequenceID(id uint32) {
	d.Header.Sequence = id
}
