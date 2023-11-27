package smpp50

import (
	"github.com/hujm2023/go-sms-protocol/packet"
	"github.com/hujm2023/go-sms-protocol/smpp"
)

type SubmitSm struct {
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

	// CString, 1 or max 17
	ScheduleDeliveryTime string

	// CString, 1 or max 17
	ValidityPeriod string

	RegisteredDelivery   uint8
	ReplaceIfPresentFlag uint8
	DataCoding           uint8
	SmDefaultMsgID       uint8
	SmLength             uint8

	// String, max 255
	ShortMessage []byte

	tlv smpp.TLVs
}

func (s *SubmitSm) IDecode(data []byte) error {
	buf := packet.NewPacketReader(data)
	defer buf.Release()

	s.Header = smpp.ReadHeader(buf)
	s.ServiceType = buf.ReadCString()
	s.SourceAddrTon = buf.ReadUint8()
	s.SourceAddrNpi = buf.ReadUint8()
	s.SourceAddr = buf.ReadCString()
	s.DestAddrTon = buf.ReadUint8()
	s.DestAddrNpi = buf.ReadUint8()
	s.DestinationAddr = buf.ReadCString()
	s.ESMClass = buf.ReadUint8()
	s.ProtocolID = buf.ReadUint8()
	s.PriorityFlag = buf.ReadUint8()
	s.ScheduleDeliveryTime = buf.ReadCString()
	s.ValidityPeriod = buf.ReadCString()
	s.RegisteredDelivery = buf.ReadUint8()
	s.ReplaceIfPresentFlag = buf.ReadUint8()
	s.DataCoding = buf.ReadUint8()
	s.SmDefaultMsgID = buf.ReadUint8()
	s.SmLength = buf.ReadUint8()
	s.ShortMessage = buf.ReadNBytes(int(s.SmLength))
	s.tlv = smpp.ReadTLVs1(buf)

	return buf.Error()
}

func (s *SubmitSm) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter()
	defer buf.Release()

	smpp.WriteHeaderNoLength(s.Header, buf)
	buf.WriteCString(s.ServiceType)
	buf.WriteUint8(s.SourceAddrTon)
	buf.WriteUint8(s.SourceAddrNpi)
	buf.WriteCString(s.SourceAddr)
	buf.WriteUint8(s.DestAddrTon)
	buf.WriteUint8(s.DestAddrNpi)
	buf.WriteCString(s.DestinationAddr)
	buf.WriteUint8(s.ESMClass)
	buf.WriteUint8(s.ProtocolID)
	buf.WriteUint8(s.PriorityFlag)
	buf.WriteCString(s.ScheduleDeliveryTime)
	buf.WriteCString(s.ValidityPeriod)
	buf.WriteUint8(s.RegisteredDelivery)
	buf.WriteUint8(s.ReplaceIfPresentFlag)
	buf.WriteUint8(s.DataCoding)
	buf.WriteUint8(s.SmDefaultMsgID)
	buf.WriteUint8(s.SmLength)
	buf.WriteBytes(s.ShortMessage)
	buf.WriteBytes(s.tlv.Bytes())

	return buf.BytesWithLength()
}

func (s *SubmitSm) SetSequenceID(id uint32) {
	s.Header.Sequence = id
}

type SubmitSmResp struct {
	smpp.Header

	// CString, max 65
	MessageID string

	tlv smpp.TLVs
}

func (s *SubmitSmResp) IDecode(data []byte) error {
	buf := packet.NewPacketReader(data)
	defer buf.Release()

	s.Header = smpp.ReadHeader(buf)
	s.MessageID = buf.ReadCString()
	s.tlv = smpp.ReadTLVs1(buf)

	return buf.Error()
}

func (s *SubmitSmResp) IEncode() ([]byte, error) {
	buf := packet.NewPacketWriter()
	defer buf.Release()

	buf.WriteUint32(uint32(s.Header.ID))
	buf.WriteUint32(uint32(s.Header.Status))
	buf.WriteUint32(s.Sequence)
	buf.WriteCString(s.MessageID)
	buf.WriteBytes(s.tlv.Bytes())

	return buf.BytesWithLength()
}

func (s *SubmitSmResp) SetSequenceID(id uint32) {
	s.Header.Sequence = id
}
