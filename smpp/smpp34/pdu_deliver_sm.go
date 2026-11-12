package smpp34

import (
	"fmt"

	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/packet"
	"github.com/hujm2023/go-sms-protocol/smpp"
)

type DeliverSm struct {
	smpp.Header

	// CString; length is bounded by the total PDU size.
	ServiceType string

	SourceAddrTon uint8
	SourceAddrNpi uint8
	// CString; length is bounded by the total PDU size.
	SourceAddr string

	DestAddrTon uint8
	DestAddrNpi uint8
	// CString; length is bounded by the total PDU size.
	DestinationAddr string

	ESMClass     uint8
	ProtocolID   uint8
	PriorityFlag uint8

	// CString; length is bounded by the total PDU size.
	ScheduleDeliveryTime string

	// CString; length is bounded by the total PDU size.
	ValidityPeriod string

	RegisteredDelivery   uint8
	ReplaceIfPresentFlag uint8
	DataCoding           uint8
	SmDefaultMsgId       uint8

	SmLength     uint8
	ShortMessage []byte

	// OptionalTLVs preserves wire order and repeated tags. When non-nil it is
	// used for encoding; TLVs remains the legacy last-value-by-tag view.
	OptionalTLVs smpp.TLVList
	TLVs         smpp.TLVs
}

func (d *DeliverSm) IDecode(data []byte) error {
	header, err := smpp.ValidateDecodedPDU(data, smpp.DELIVER_SM, false)
	if err != nil {
		return err
	}
	if err := smpp.ValidateRequestHeader(header, smpp.DELIVER_SM); err != nil {
		return err
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
	d.ShortMessage = buf.ReadNBytes(int(d.SmLength))
	d.OptionalTLVs, err = smpp.ReadTLVList(buf)
	if err != nil {
		return err
	}
	d.TLVs = d.OptionalTLVs.ToMap()

	if err := buf.Error(); err != nil {
		return err
	}
	return d.validate()
}

func (d *DeliverSm) IEncode() ([]byte, error) {
	if err := d.validate(); err != nil {
		return nil, err
	}
	tlvBytes, err := d.optionalTLVs().MarshalBinary()
	if err != nil {
		return nil, err
	}

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
	buf.WriteBytes(tlvBytes)

	data, err := buf.BytesWithLength()
	if err != nil {
		return nil, err
	}
	if err := smpp.ValidateEncodedPDU(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (d *DeliverSm) optionalTLVs() smpp.TLVList {
	if d.OptionalTLVs != nil {
		return d.OptionalTLVs
	}
	return d.TLVs.ToList()
}

func (d *DeliverSm) validate() error {
	if err := smpp.ValidateRequestHeader(d.Header, smpp.DELIVER_SM); err != nil {
		return err
	}
	if err := validateMessageAddressFields(d.ServiceType, d.SourceAddrTon, d.SourceAddrNpi, d.SourceAddr, d.DestAddrTon, d.DestAddrNpi, d.DestinationAddr); err != nil {
		return err
	}
	switch messageType := d.ESMClass & 0x3c; messageType {
	case 0x00, 0x04, 0x08, 0x10, 0x18, 0x20:
	default:
		return fmt.Errorf("deliver_sm esm_class message type %#02x is reserved", messageType)
	}
	if err := smpp.ValidatePriorityFlag(d.PriorityFlag); err != nil {
		return err
	}
	if err := smpp.ValidateCString(smpp.SCHEDULE_DELIVERY_TIME, d.ScheduleDeliveryTime, smpp.MAX_PDU_SIZE); err != nil {
		return err
	}
	if err := smpp.ValidateCString(smpp.VALIDITY_PERIOD, d.ValidityPeriod, smpp.MAX_PDU_SIZE); err != nil {
		return err
	}
	if d.ReplaceIfPresentFlag != 0 {
		return fmt.Errorf("deliver_sm replace_if_present_flag must be zero")
	}
	if d.SmDefaultMsgId != 0 {
		return fmt.Errorf("deliver_sm sm_default_msg_id must be zero")
	}
	if int(d.SmLength) != len(d.ShortMessage) {
		return fmt.Errorf("sm_length=%d does not match short_message length=%d", d.SmLength, len(d.ShortMessage))
	}

	return smpp.ValidateTLVList(d.optionalTLVs())
}

func (d *DeliverSm) SetSequenceID(id uint32) {
	d.Header.Sequence = id
}

func (d *DeliverSm) GetSequenceID() uint32 {
	return d.Header.Sequence
}

func (d *DeliverSm) GetCommand() sms.ICommander {
	return smpp.DELIVER_SM
}

func (d *DeliverSm) GenEmptyResponse() sms.PDU {
	return &DeliverSmResp{
		Header: smpp.Header{
			ID:       smpp.DELIVER_SM_RESP,
			Sequence: d.Header.Sequence,
		},
	}
}

func (d *DeliverSm) String() string {
	s := packet.NewPDUStringer()
	defer s.Release()

	s.Write("Header", d.Header)
	s.Write("ServiceType", d.ServiceType)
	s.Write("SourceAddrTon", d.SourceAddrTon)
	s.Write("SourceAddrNpi", d.SourceAddrNpi)
	s.Write("SourceAddr", d.SourceAddr)
	s.Write("DestAddrTon", d.DestAddrTon)
	s.Write("DestAddrNpi", d.DestAddrNpi)
	s.Write("DestinationAddr", d.DestinationAddr)
	s.Write("ESMClass", d.ESMClass)
	s.Write("ProtocolID", d.ProtocolID)
	s.Write("PriorityFlag", d.PriorityFlag)
	s.Write("ScheduleDeliveryTime", d.ScheduleDeliveryTime)
	s.Write("ValidityPeriod", d.ValidityPeriod)
	s.Write("RegisteredDelivery", d.RegisteredDelivery)
	s.Write("ReplaceIfPresentFlag", d.ReplaceIfPresentFlag)
	s.Write("DataCoding", d.DataCoding)
	s.Write("SmDefaultMsgId", d.SmDefaultMsgId)
	s.Write("SmLength", d.SmLength)
	s.Write("ShortMessage", d.ShortMessage)
	if d.OptionalTLVs != nil {
		s.OmitWrite("TLVs", d.OptionalTLVs.String())
	} else {
		s.OmitWrite("TLVs", d.TLVs.String())
	}

	return s.String()
}

type DeliverSmResp struct {
	smpp.Header

	// CString; length is bounded by the total PDU size.
	MessageID string
}

func (d *DeliverSmResp) IDecode(data []byte) error {
	header, err := smpp.ValidateDecodedPDU(data, smpp.DELIVER_SM_RESP, false)
	if err != nil {
		return err
	}
	if err := smpp.ValidateResponseHeader(header, smpp.DELIVER_SM_RESP); err != nil {
		return err
	}
	buf := packet.NewPacketReader(data)
	defer buf.Release()

	d.Header = smpp.ReadHeader(buf)
	d.MessageID = buf.ReadCString()
	if err := buf.Error(); err != nil {
		return err
	}
	if buf.Remaining() != 0 {
		return fmt.Errorf("deliver_sm_resp contains %d trailing octets", buf.Remaining())
	}
	return smpp.ValidateCString(smpp.MESSAGE_ID, d.MessageID, smpp.MAX_PDU_SIZE)
}

func (d *DeliverSmResp) IEncode() ([]byte, error) {
	if err := smpp.ValidateResponseHeader(d.Header, smpp.DELIVER_SM_RESP); err != nil {
		return nil, err
	}
	if err := smpp.ValidateCString(smpp.MESSAGE_ID, d.MessageID, smpp.MAX_PDU_SIZE); err != nil {
		return nil, err
	}

	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	smpp.WriteHeaderNoLength(d.Header, buf)

	buf.WriteCString(d.MessageID)

	data, err := buf.BytesWithLength()
	if err != nil {
		return nil, err
	}
	if err := smpp.ValidateEncodedPDU(data); err != nil {
		return nil, err
	}
	return data, nil
}

func (d *DeliverSmResp) SetSequenceID(id uint32) {
	d.Header.Sequence = id
}

func (d *DeliverSmResp) GetSequenceID() uint32 {
	return d.Header.Sequence
}

func (d *DeliverSmResp) GetCommand() sms.ICommander {
	return smpp.DELIVER_SM_RESP
}

func (d *DeliverSmResp) GenEmptyResponse() sms.PDU {
	return nil
}

func (d *DeliverSmResp) String() string {
	s := packet.NewPDUStringer()
	defer s.Release()

	s.Write("Header", d.Header)
	s.Write("MessageID", d.MessageID)

	return s.String()
}
