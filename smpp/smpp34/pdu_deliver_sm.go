package smpp34

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
	// CString, max 21
	SourceAddr string

	DestAddrTon uint8
	DestAddrNpi uint8
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
	SmDefaultMsgId       uint8

	SmLength     uint8
	ShortMessage []byte

	tlvs smpp.TLVs
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
	d.SmDefaultMsgId = buf.ReadUint8()
	d.SmLength = buf.ReadUint8()
	temp := make([]byte, d.SmLength)
	buf.ReadBytes(temp)
	d.ShortMessage = temp
	d.tlvs = smpp.ReadTLVs1(buf)

	return buf.Error()
}

func (d *DeliverSm) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter(0)
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
	buf.WriteUint8(d.SmDefaultMsgId)
	buf.WriteUint8(d.SmLength)
	buf.WriteBytes(d.ShortMessage)
	buf.WriteBytes(d.tlvs.Bytes())

	return buf.BytesWithLength()
}

func (d *DeliverSm) SetSequenceID(id uint32) {
	d.Header.Sequence = id
}

type DeliverSmResp struct {
	smpp.Header

	// CString, size 1, unused, set to null
	MessageID string
}

func (d *DeliverSmResp) IDecode(data []byte) error {
	if len(data) < smpp.MinSMPPPacketLen {
		return smpp.ErrInvalidPudLength
	}
	buf := packet.NewPacketReader(data)
	defer buf.Release()

	d.Header = smpp.ReadHeader(buf)
	d.MessageID = buf.ReadCString()
	return buf.Error()
}

func (d *DeliverSmResp) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	smpp.WriteHeaderNoLength(d.Header, buf)

	buf.WriteCString(d.MessageID)

	return buf.BytesWithLength()
}

func (d *DeliverSmResp) SetSequenceID(id uint32) {
	d.Header.Sequence = id
}
