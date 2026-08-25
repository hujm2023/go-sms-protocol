package smpp34

import (
	"fmt"

	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/packet"
	"github.com/hujm2023/go-sms-protocol/smpp"
)

type Bind struct {
	smpp.Header

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

func (b *Bind) GetSequenceID() uint32 {
	return b.Header.Sequence
}

func (b *Bind) GetCommand() sms.ICommander {
	switch b.Header.ID {
	case smpp.BIND_RECEIVER, smpp.BIND_TRANSMITTER, smpp.BIND_TRANSCEIVER:
		return b.Header.ID
	default:
		return nil
	}
}

func (b *Bind) GenEmptyResponse() sms.PDU {
	var responseID smpp.CMDId
	switch b.Header.ID {
	case smpp.BIND_RECEIVER:
		responseID = smpp.BIND_RECEIVER_RESP
	case smpp.BIND_TRANSMITTER:
		responseID = smpp.BIND_TRANSMITTER_RESP
	case smpp.BIND_TRANSCEIVER:
		responseID = smpp.BIND_TRANSCEIVER_RESP
	default:
		return nil
	}

	return &BindResp{
		Header: smpp.Header{
			ID:       responseID,
			Sequence: b.Header.Sequence,
		},
	}
}

func (b *Bind) IDecode(data []byte) error {
	header, err := smpp.PeekHeader(data)
	if err != nil {
		return err
	}
	if !isBindRequestID(header.ID) {
		return fmt.Errorf("unexpected bind command_id %s", header.ID)
	}
	header, err = smpp.ValidateDecodedPDU(data, header.ID, false)
	if err != nil {
		return err
	}
	if err := smpp.ValidateRequestHeader(header, header.ID); err != nil {
		return err
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	b.Header = smpp.ReadHeader(buf)
	b.SystemID = buf.ReadCString()
	b.Password = buf.ReadCString()
	b.SystemType = buf.ReadCString()
	b.InterfaceVersion = buf.ReadUint8()
	b.AddrTon = buf.ReadUint8()
	b.AddrNpi = buf.ReadUint8()
	b.AddressRange = buf.ReadCString()

	if err := buf.Error(); err != nil {
		return err
	}
	if buf.Remaining() != 0 {
		return fmt.Errorf("bind PDU contains %d trailing octets", buf.Remaining())
	}
	return b.validate()
}

func (b *Bind) IEncode() ([]byte, error) {
	if err := b.validate(); err != nil {
		return nil, err
	}
	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	smpp.WriteHeaderNoLength(b.Header, buf)

	buf.WriteCString(b.SystemID)
	buf.WriteCString(b.Password)
	buf.WriteCString(b.SystemType)
	buf.WriteUint8(b.InterfaceVersion)
	buf.WriteUint8(b.AddrTon)
	buf.WriteUint8(b.AddrNpi)
	buf.WriteCString(b.AddressRange)

	data, err := buf.BytesWithLength()
	if err != nil {
		return nil, err
	}
	return data, smpp.ValidateEncodedPDU(data)
}

func (b *Bind) validate() error {
	if !isBindRequestID(b.Header.ID) {
		return fmt.Errorf("unexpected bind command_id %s", b.Header.ID)
	}
	if err := smpp.ValidateRequestHeader(b.Header, b.Header.ID); err != nil {
		return err
	}
	if err := smpp.ValidateCString(smpp.SYSTEM_ID, b.SystemID, 16); err != nil {
		return err
	}
	if err := smpp.ValidateCString(smpp.PASSWORD, b.Password, 9); err != nil {
		return err
	}
	if err := smpp.ValidateCString(smpp.SYSTEM_TYPE, b.SystemType, 13); err != nil {
		return err
	}
	if err := smpp.ValidateInterfaceVersion(b.InterfaceVersion); err != nil {
		return err
	}
	if err := smpp.ValidateTON(smpp.ADDR_TON, b.AddrTon); err != nil {
		return err
	}
	if err := smpp.ValidateNPI(smpp.ADDR_NPI, b.AddrNpi); err != nil {
		return err
	}
	return smpp.ValidateCString(smpp.ADDRESS_RANGE, b.AddressRange, 41)
}

func isBindRequestID(id smpp.CMDId) bool {
	switch id {
	case smpp.BIND_RECEIVER, smpp.BIND_TRANSMITTER, smpp.BIND_TRANSCEIVER:
		return true
	default:
		return false
	}
}

func (b *Bind) SetSequenceID(id uint32) {
	b.Header.Sequence = id
}

func (b *Bind) String() string {
	s := packet.NewPDUStringer()
	defer s.Release()

	s.Write("Header", b.Header)
	s.Write("SystemID", b.SystemID)
	s.Write("Password", b.Password)
	s.Write("SystemType", b.SystemType)
	s.Write("InterfaceVersion", b.InterfaceVersion)
	s.Write("AddrTon", b.AddrTon)
	s.Write("AddrNpi", b.AddrNpi)
	s.Write("AddressRange", b.AddressRange)

	return s.String()
}

type BindResp struct {
	smpp.Header

	// CString, max 16
	SystemID string

	TLVs smpp.TLVs
}

func (b *BindResp) GetSequenceID() uint32 {
	return b.Header.Sequence
}

func (b *BindResp) GetCommand() sms.ICommander {
	switch b.Header.ID {
	case smpp.BIND_RECEIVER_RESP, smpp.BIND_TRANSMITTER_RESP, smpp.BIND_TRANSCEIVER_RESP:
		return b.Header.ID
	default:
		return nil
	}
}

func (b *BindResp) GenEmptyResponse() sms.PDU {
	return nil
}

func (b *BindResp) IEncode() ([]byte, error) {
	if !isBindResponseID(b.Header.ID) {
		return nil, fmt.Errorf("unexpected bind_resp command_id %s", b.Header.ID)
	}
	if err := smpp.ValidateResponseHeader(b.Header, b.Header.ID); err != nil {
		return nil, err
	}
	var tlvBytes []byte
	if b.Header.Status == smpp.ESME_ROK {
		if err := smpp.ValidateCString(smpp.SYSTEM_ID, b.SystemID, 16); err != nil {
			return nil, err
		}
		tlvList := b.TLVs.ToList()
		if err := smpp.ValidateTLVList(tlvList); err != nil {
			return nil, err
		}
		var err error
		tlvBytes, err = tlvList.MarshalBinary()
		if err != nil {
			return nil, err
		}
	}

	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	smpp.WriteHeaderNoLength(b.Header, buf)

	if b.Header.Status != smpp.ESME_ROK {
		data, err := buf.BytesWithLength()
		if err != nil {
			return nil, err
		}
		return data, smpp.ValidateEncodedPDU(data)
	}

	buf.WriteCString(b.SystemID)

	buf.WriteBytes(tlvBytes)

	data, err := buf.BytesWithLength()
	if err != nil {
		return nil, err
	}
	return data, smpp.ValidateEncodedPDU(data)
}

func (b *BindResp) IDecode(data []byte) error {
	header, err := smpp.PeekHeader(data)
	if err != nil {
		return err
	}
	if !isBindResponseID(header.ID) {
		return fmt.Errorf("unexpected bind_resp command_id %s", header.ID)
	}
	header, err = smpp.ValidateDecodedPDU(data, header.ID, false)
	if err != nil {
		return err
	}
	if err := smpp.ValidateResponseHeader(header, header.ID); err != nil {
		return err
	}

	buf := packet.NewPacketReader(data)
	defer buf.Release()

	b.Header = smpp.ReadHeader(buf)

	if b.Header.Status != smpp.ESME_ROK {
		b.SystemID = ""
		b.TLVs = nil
		if len(data) != smpp.MinSMPPPacketLen {
			return fmt.Errorf("bind_resp with error status must not contain a body")
		}
		return nil
	}

	b.SystemID = buf.ReadCString()

	tlvList, err := smpp.ReadTLVList(buf)
	if err != nil {
		return err
	}
	b.TLVs = tlvList.ToMap()

	if err := buf.Error(); err != nil {
		return err
	}
	if err := smpp.ValidateCString(smpp.SYSTEM_ID, b.SystemID, 16); err != nil {
		return err
	}
	return smpp.ValidateTLVList(tlvList)
}

func isBindResponseID(id smpp.CMDId) bool {
	switch id {
	case smpp.BIND_RECEIVER_RESP, smpp.BIND_TRANSMITTER_RESP, smpp.BIND_TRANSCEIVER_RESP:
		return true
	default:
		return false
	}
}

func (b *BindResp) SetSequenceID(id uint32) {
	b.Header.Sequence = id
}

func (b *BindResp) String() string {
	s := packet.NewPDUStringer()
	defer s.Release()

	s.Write("Header", b.Header)
	s.Write("SystemID", b.SystemID)
	s.OmitWrite("TLVs", b.TLVs.String())

	return s.String()
}
