package smpp34

import (
	"fmt"

	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/packet"
	"github.com/hujm2023/go-sms-protocol/smpp"
)

type Unbind struct {
	smpp.Header
}

func (u *Unbind) IDecode(data []byte) error {
	header, err := smpp.ValidateDecodedPDU(data, smpp.UNBIND, false)
	if err != nil {
		return err
	}
	if len(data) != smpp.MinSMPPPacketLen {
		return fmt.Errorf("unbind must not contain a body")
	}
	if err := smpp.ValidateRequestHeader(header, smpp.UNBIND); err != nil {
		return err
	}
	u.Header = header
	return nil
}

func (u *Unbind) IEncode() ([]byte, error) {
	if err := smpp.ValidateRequestHeader(u.Header, smpp.UNBIND); err != nil {
		return nil, err
	}
	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	smpp.WriteHeaderNoLength(u.Header, buf)

	data, err := buf.BytesWithLength()
	if err != nil {
		return nil, err
	}
	return data, smpp.ValidateEncodedPDU(data)
}

func (u *Unbind) SetSequenceID(id uint32) {
	u.Header.Sequence = id
}

func (u *Unbind) GetSequenceID() uint32 {
	return u.Header.Sequence
}

func (u *Unbind) GetCommand() sms.ICommander {
	return smpp.UNBIND
}

func (u *Unbind) GenEmptyResponse() sms.PDU {
	return &UnBindResp{
		Header: smpp.Header{
			ID:       smpp.UNBIND_RESP,
			Sequence: u.Header.Sequence,
		},
	}
}

type UnBindResp struct {
	smpp.Header
}

func (u *UnBindResp) IDecode(data []byte) error {
	header, err := smpp.ValidateDecodedPDU(data, smpp.UNBIND_RESP, false)
	if err != nil {
		return err
	}
	if len(data) != smpp.MinSMPPPacketLen {
		return fmt.Errorf("unbind_resp must not contain a body")
	}
	if err := smpp.ValidateResponseHeader(header, smpp.UNBIND_RESP); err != nil {
		return err
	}
	u.Header = header
	return nil
}

func (u *UnBindResp) IEncode() ([]byte, error) {
	if err := smpp.ValidateResponseHeader(u.Header, smpp.UNBIND_RESP); err != nil {
		return nil, err
	}
	buf := packet.NewPacketWriter(0)
	defer buf.Release()

	smpp.WriteHeaderNoLength(u.Header, buf)

	data, err := buf.BytesWithLength()
	if err != nil {
		return nil, err
	}
	return data, smpp.ValidateEncodedPDU(data)
}

func (u *UnBindResp) SetSequenceID(id uint32) {
	u.Header.Sequence = id
}

func (u *UnBindResp) GetSequenceID() uint32 {
	return u.Header.Sequence
}

func (u *UnBindResp) GetCommand() sms.ICommander {
	return smpp.UNBIND_RESP
}

func (u *UnBindResp) GenEmptyResponse() sms.PDU {
	return nil
}

func (u *UnBindResp) String() string {
	str := packet.NewPDUStringer()
	defer str.Release()

	str.Write("Header", u.Header)

	return str.String()
}
