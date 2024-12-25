package smpp34

import (
	sms "github.com/hujm2023/go-sms-protocol"
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

	SmLength uint8
	// Uint8, max 254
	ShortMessage []byte

	TLVs smpp.TLVs
}

func (s *SubmitSm) IDecode(data []byte) error {
	if len(data) < smpp.MinSMPPPacketLen {
		return smpp.ErrInvalidPudLength
	}

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
	s.ShortMessage = r.ReadNBytes(int(s.SmLength))
	s.TLVs = smpp.ReadTLVs1(r)

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
	w.WriteBytes(s.TLVs.Bytes())

	return w.BytesWithLength()
}

func (s *SubmitSm) SetSequenceID(id uint32) {
	s.Header.Sequence = id
}

func (s *SubmitSm) GetSequenceID() uint32 {
	return s.Header.Sequence
}

func (s *SubmitSm) GetCommand() sms.ICommander {
	return smpp.SUBMIT_SM
}

func (s *SubmitSm) GenEmptyResponse() sms.PDU {
	return &SubmitSmResp{
		Header: smpp.Header{
			ID:       smpp.SUBMIT_SM_RESP,
			Sequence: s.Header.Sequence,
		},
	}
}

func (s *SubmitSm) String() string {
	str := packet.NewPDUStringer()
	defer str.Release()

	str.Write("Header", s.Header)
	str.Write("ServiceType", s.ServiceType)
	str.Write("SourceAddrTon", s.SourceAddrTon)
	str.Write("SourceAddrNpi", s.SourceAddrNpi)
	str.Write("SourceAddr", s.SourceAddr)
	str.Write("DestAddrTon", s.DestAddrTon)
	str.Write("DestAddrNpi", s.DestAddrNpi)
	str.Write("DestinationAddr", s.DestinationAddr)
	str.Write("ESMClass", s.ESMClass)
	str.Write("ProtocolID", s.ProtocolID)
	str.Write("PriorityFlag", s.PriorityFlag)
	str.Write("ScheduleDeliveryTime", s.ScheduleDeliveryTime)
	str.Write("ValidityPeriod", s.ValidityPeriod)
	str.Write("RegisteredDelivery", s.RegisteredDelivery)
	str.Write("ReplaceIfPresentFlag", s.ReplaceIfPresentFlag)
	str.Write("DataCoding", s.DataCoding)
	str.Write("SmDefaultMsgID", s.SmDefaultMsgID)
	str.Write("SmLength", s.SmLength)
	str.Write("ShortMessage", s.ShortMessage)
	str.OmitWrite("TLVs", s.TLVs.String())

	return str.String()
}

type SubmitSmResp struct {
	smpp.Header
	// CString, max 65
	MessageID string
}

func (s *SubmitSmResp) GetSequenceID() uint32 {
	return s.Header.Sequence
}

func (s *SubmitSmResp) GetCommand() sms.ICommander {
	return smpp.SUBMIT_SM_RESP
}

func (s *SubmitSmResp) GenEmptyResponse() sms.PDU {
	return nil
}

func (s *SubmitSmResp) IDecode(data []byte) error {
	if len(data) < smpp.MinSMPPPacketLen {
		return smpp.ErrInvalidPudLength
	}

	r := packet.NewPacketReader(data)
	defer r.Release()

	s.Header = smpp.ReadHeader(r)
	s.MessageID = r.ReadCString()

	return r.Error()
}

func (s *SubmitSmResp) IEncode() ([]byte, error) {
	w := packet.NewPacketWriter(0)
	defer w.Release()

	smpp.WriteHeaderNoLength(s.Header, w)
	w.WriteCString(s.MessageID)

	return w.BytesWithLength()
}

func (s *SubmitSmResp) SetSequenceID(id uint32) {
	s.Header.Sequence = id
}

func (s *SubmitSmResp) String() string {
	str := packet.NewPDUStringer()
	defer str.Release()

	str.Write("Header", s.Header)
	str.Write("MessageID", s.MessageID)

	return str.String()
}
