package cmpp30

import (
	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/packet"
)

type Terminate struct {
	cmpp.Header
}

func (t *Terminate) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return cmpp.ErrInvalidPudLength
	}
	b := packet.NewPacketReader(data)
	defer b.Release()

	t.Header = cmpp.ReadHeader(b)
	return b.Error()
}

func (t *Terminate) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()

	cmpp.WriteHeaderNoLength(t.Header, b)

	return b.BytesWithLength()
}

func (t *Terminate) SetSequenceID(id uint32) {
	t.Header.SequenceID = id
}

func (t *Terminate) GetSequenceID() uint32 {
	return t.Header.SequenceID
}

func (t *Terminate) GetCommand() sms.ICommander {
	return cmpp.CommandTerminate
}

func (t *Terminate) GenEmptyResponse() sms.PDU {
	return &TerminateResp{
		Header: cmpp.Header{
			CommandID:  cmpp.CommandTerminateResp,
			SequenceID: t.GetSequenceID(),
		},
	}
}

func (t *Terminate) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", t.Header)

	return w.String()
}

type TerminateResp struct {
	Header cmpp.Header
}

func (t *TerminateResp) IDecode(data []byte) error {
	if len(data) < cmpp.MinCMPPPduLength {
		return cmpp.ErrInvalidPudLength
	}
	b := packet.NewPacketReader(data)
	defer b.Release()

	t.Header = cmpp.ReadHeader(b)
	return b.Error()
}

func (t *TerminateResp) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()

	cmpp.WriteHeaderNoLength(t.Header, b)

	return b.BytesWithLength()
}

func (t *TerminateResp) SetSequenceID(id uint32) {
	t.Header.SequenceID = id
}

func (t *TerminateResp) GetSequenceID() uint32 {
	return t.Header.SequenceID
}

func (t *TerminateResp) GetCommand() sms.ICommander {
	return cmpp.CommandTerminateResp
}

func (t *TerminateResp) GenEmptyResponse() sms.PDU {
	return nil
}

func (t *TerminateResp) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", t.Header)

	return w.String()
}
