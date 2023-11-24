package cmpp30

import (
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

	b.WriteUint32(uint32(t.Header.CommandID))
	b.WriteUint32(t.Header.SequenceID)
	return b.BytesWithLength()
}

func (t *Terminate) SetSequenceID(id uint32) {
	t.Header.SequenceID = id
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

	b.WriteUint32(uint32(t.Header.CommandID))
	b.WriteUint32(t.Header.SequenceID)
	return b.BytesWithLength()
}

func (t *TerminateResp) SetSequenceID(id uint32) {
	t.Header.SequenceID = id
}
