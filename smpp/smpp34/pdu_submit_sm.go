package smpp34

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

	// CString, 1~17
	ScheduleDeliveryTime string
	// CString, 1~17
	ValidityPeriod string

	RegisteredDelivery   uint8
	ReplaceIfPresentFlag uint8
	DataCoding           uint8
	SmDefaultMsgID       uint8
	SmLength             uint8

	// Uint8, max 254
	ShortMessage []byte

	tlvs smpp.TLVs
}

func (s *SubmitSm) IDecode(data []byte) error {
	r := packet.NewPacketReader(data)
	defer r.Release()

	s.Header = smpp.ReadHeader(r)
	s.ServiceType = r.ReadCString()
	s.SourceAddrTon = r.ReadUint8()
	s.SourceAddrNpi = r.ReadUint8()
	s.SourceAddr = r.ReadCString()
	s.DestAddrTon = r.ReadUint8()
	s.DestAddrNpi = r.ReadUint8()
	s.DestinationAddr = r.ReadCString()
	s.ESMClass = r.ReadUint8()
	s.ProtocolID = r.ReadUint8()
	s.PriorityFlag = r.ReadUint8()
	s.ScheduleDeliveryTime = r.ReadCString()
	s.ValidityPeriod = r.ReadCString()
	s.RegisteredDelivery = r.ReadUint8()
	s.ReplaceIfPresentFlag = r.ReadUint8()
	s.DataCoding = r.ReadUint8()
	s.SmDefaultMsgID = r.ReadUint8()
	s.SmLength = r.ReadUint8()
	temp := make([]byte, s.SmLength)
	r.ReadBytes(temp)
	s.ShortMessage = temp
	s.tlvs = smpp.ReadTLVs1(r)

	return r.Error()
}

func (s *SubmitSm) IEncode() ([]byte, error) {
	w := packet.NewPacketWriter(0)
	defer w.Release()

	smpp.WriteHeaderNoLength(s.Header, w)

	w.WriteCString(s.ServiceType)
	w.WriteUint8(s.SourceAddrTon)
	w.WriteUint8(s.SourceAddrNpi)
	w.WriteCString(s.SourceAddr)
	w.WriteUint8(s.DestAddrTon)
	w.WriteUint8(s.DestAddrNpi)
	w.WriteCString(s.DestinationAddr)
	w.WriteUint8(s.ESMClass)
	w.WriteUint8(s.ProtocolID)
	w.WriteUint8(s.PriorityFlag)
	w.WriteCString(s.ScheduleDeliveryTime)
	w.WriteCString(s.ValidityPeriod)
	w.WriteUint8(s.RegisteredDelivery)
	w.WriteUint8(s.ReplaceIfPresentFlag)
	w.WriteUint8(s.DataCoding)
	w.WriteUint8(s.SmDefaultMsgID)
	w.WriteUint8(s.SmLength)
	w.WriteBytes(s.ShortMessage)
	w.WriteBytes(s.tlvs.Bytes())

	return w.BytesWithLength()
}

func (s *SubmitSm) SetSequenceID(id uint32) {
	s.Header.Sequence = id
}

type SubmitSMResp struct {
	smpp.Header
	// CString, max 65
	MessageID string
}

func (s *SubmitSMResp) IDecode(data []byte) error {
	r := packet.NewPacketReader(data)
	defer r.Release()

	s.Header = smpp.ReadHeader(r)
	s.MessageID = r.ReadCString()

	return r.Error()
}

func (s *SubmitSMResp) IEncode() ([]byte, error) {
	w := packet.NewPacketWriter(0)
	defer w.Release()

	smpp.WriteHeaderNoLength(s.Header, w)
	w.WriteCString(s.MessageID)

	return w.BytesWithLength()
}

func (s *SubmitSMResp) SetSequenceID(id uint32) {
	s.Header.Sequence = id
}
