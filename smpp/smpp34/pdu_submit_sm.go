package smpp34

import (
	"fmt"

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

	// OptionalTLVs preserves wire order and repeated tags. When non-nil it is
	// used for encoding; TLVs remains the legacy last-value-by-tag view.
	OptionalTLVs smpp.TLVList
	TLVs         smpp.TLVs
}

func (s *SubmitSm) IDecode(data []byte) error {
	header, err := smpp.ValidateDecodedPDU(data, smpp.SUBMIT_SM, false)
	if err != nil {
		return err
	}
	if err := smpp.ValidateRequestHeader(header, smpp.SUBMIT_SM); err != nil {
		return err
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
	s.OptionalTLVs, err = smpp.ReadTLVList(r)
	if err != nil {
		return err
	}
	s.TLVs = s.OptionalTLVs.ToMap()

	if err := r.Error(); err != nil {
		return err
	}
	return s.validate()
}

func (s *SubmitSm) IEncode() ([]byte, error) {
	if err := s.validate(); err != nil {
		return nil, err
	}
	tlvBytes, err := s.optionalTLVs().MarshalBinary()
	if err != nil {
		return nil, err
	}

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
	w.WriteBytes(tlvBytes)

	data, err := w.BytesWithLength()
	if err != nil {
		return nil, err
	}
	if err := smpp.ValidateEncodedPDU(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (s *SubmitSm) optionalTLVs() smpp.TLVList {
	if s.OptionalTLVs != nil {
		return s.OptionalTLVs
	}
	return s.TLVs.ToList()
}

func (s *SubmitSm) validate() error {
	if err := smpp.ValidateRequestHeader(s.Header, smpp.SUBMIT_SM); err != nil {
		return err
	}
	if err := validateMessageAddressFields(s.ServiceType, s.SourceAddrTon, s.SourceAddrNpi, s.SourceAddr, s.DestAddrTon, s.DestAddrNpi, s.DestinationAddr); err != nil {
		return err
	}
	if messageType := s.ESMClass & 0x3c; messageType != 0 && messageType != 0x08 && messageType != 0x10 {
		return fmt.Errorf("submit_sm esm_class message type %#02x is reserved", messageType)
	}
	if err := smpp.ValidatePriorityFlag(s.PriorityFlag); err != nil {
		return err
	}
	if err := smpp.ValidateTimeCString(smpp.SCHEDULE_DELIVERY_TIME, s.ScheduleDeliveryTime); err != nil {
		return err
	}
	if err := smpp.ValidateTimeCString(smpp.VALIDITY_PERIOD, s.ValidityPeriod); err != nil {
		return err
	}
	if s.RegisteredDelivery&^uint8(0x2f) != 0 || s.RegisteredDelivery&0x03 == 0x03 {
		return fmt.Errorf("submit_sm registered_delivery %#02x uses reserved bits or value", s.RegisteredDelivery)
	}
	if s.ReplaceIfPresentFlag > 1 {
		return fmt.Errorf("replace_if_present_flag must be 0 or 1, got %d", s.ReplaceIfPresentFlag)
	}
	if err := smpp.ValidateDataCoding(s.DataCoding); err != nil {
		return err
	}
	if s.SmDefaultMsgID == 0xff {
		return fmt.Errorf("sm_default_msg_id 255 is reserved")
	}
	if int(s.SmLength) != len(s.ShortMessage) || len(s.ShortMessage) > 254 {
		return fmt.Errorf("sm_length=%d does not match short_message length=%d or exceeds 254", s.SmLength, len(s.ShortMessage))
	}

	tlvs := s.optionalTLVs()
	if err := smpp.ValidateTLVList(tlvs); err != nil {
		return err
	}
	if len(tlvs.All(smpp.MESSAGE_PAYLOAD)) > 0 && (s.SmLength != 0 || len(s.ShortMessage) != 0) {
		return fmt.Errorf("message_payload and short_message are mutually exclusive")
	}
	return nil
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
	if s.OptionalTLVs != nil {
		str.OmitWrite("TLVs", s.OptionalTLVs.String())
	} else {
		str.OmitWrite("TLVs", s.TLVs.String())
	}

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
	header, err := smpp.ValidateDecodedPDU(data, smpp.SUBMIT_SM_RESP, false)
	if err != nil {
		return err
	}
	if err := smpp.ValidateResponseHeader(header, smpp.SUBMIT_SM_RESP); err != nil {
		return err
	}

	r := packet.NewPacketReader(data)
	defer r.Release()

	s.Header = smpp.ReadHeader(r)
	if s.Header.Status != smpp.ESME_ROK {
		s.MessageID = ""
		if r.Remaining() != 0 {
			return fmt.Errorf("submit_sm_resp with error status must not contain a body")
		}
		return r.Error()
	}
	s.MessageID = r.ReadCString()
	if err := r.Error(); err != nil {
		return err
	}
	if r.Remaining() != 0 {
		return fmt.Errorf("submit_sm_resp contains %d trailing octets", r.Remaining())
	}
	return smpp.ValidateCString(smpp.MESSAGE_ID, s.MessageID, 65)
}

func (s *SubmitSmResp) IEncode() ([]byte, error) {
	if err := smpp.ValidateResponseHeader(s.Header, smpp.SUBMIT_SM_RESP); err != nil {
		return nil, err
	}
	if s.Header.Status == smpp.ESME_ROK {
		if err := smpp.ValidateCString(smpp.MESSAGE_ID, s.MessageID, 65); err != nil {
			return nil, err
		}
	}

	w := packet.NewPacketWriter(0)
	defer w.Release()

	smpp.WriteHeaderNoLength(s.Header, w)
	if s.Header.Status == smpp.ESME_ROK {
		w.WriteCString(s.MessageID)
	}

	data, err := w.BytesWithLength()
	if err != nil {
		return nil, err
	}
	if err := smpp.ValidateEncodedPDU(data); err != nil {
		return nil, err
	}
	return data, nil
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
