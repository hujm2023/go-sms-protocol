package smgp30

import (
	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/packet"
	"github.com/hujm2023/go-sms-protocol/smgp"
)

type Exit struct {
	smgp.Header
}

func (t *Exit) IDecode(data []byte) error {
	if len(data) < smgp.MinSMGPPduLength {
		return smgp.ErrInvalidPudLength
	}
	b := packet.NewPacketReader(data)
	defer b.Release()

	t.Header = smgp.ReadHeader(b)
	return b.Error()
}

func (t *Exit) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()

	smgp.WriteHeaderNoLength(t.Header, b)

	return b.BytesWithLength()
}

func (t *Exit) SetSequenceID(id uint32) {
	t.Header.SequenceID = id
}

func (t *Exit) GetSequenceID() uint32 {
	return t.Header.SequenceID
}

func (t *Exit) GenerateResponseHeader() *ExitResp {
	resp := &ExitResp{
		Header: smgp.NewHeader(smgp.MaxExitRespLength, smgp.CommandExitResp, t.GetSequenceID()),
	}
	return resp
}

func (e *Exit) GetCommand() sms.ICommander {
	return smgp.CommandExit
}

func (e *Exit) GenEmptyResponse() sms.PDU {
	return &ExitResp{
		Header: smgp.Header{
			CommandID:  smgp.CommandExitResp,
			SequenceID: e.GetSequenceID(),
		},
	}
}

func (e *Exit) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", e.Header)

	return w.String()
}

type ExitResp struct {
	Header smgp.Header
}

func (t *ExitResp) IDecode(data []byte) error {
	if len(data) < smgp.MinSMGPPduLength {
		return smgp.ErrInvalidPudLength
	}
	b := packet.NewPacketReader(data)
	defer b.Release()

	t.Header = smgp.ReadHeader(b)
	return b.Error()
}

func (t *ExitResp) IEncode() ([]byte, error) {
	b := packet.NewPacketWriter()
	defer b.Release()

	smgp.WriteHeaderNoLength(t.Header, b)

	return b.BytesWithLength()
}

func (t *ExitResp) SetSequenceID(id uint32) {
	t.Header.SequenceID = id
}

func (e *ExitResp) GetSequenceID() uint32 {
	return e.Header.SequenceID
}

func (e *ExitResp) GetCommand() sms.ICommander {
	return smgp.CommandExitResp
}

func (e *ExitResp) GenEmptyResponse() sms.PDU {
	return nil
}

func (e *ExitResp) String() string {
	w := packet.NewPDUStringer()
	defer w.Release()

	w.Write("Header", e.Header)

	return w.String()
}
