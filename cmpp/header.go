package cmpp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

const (
	HeaderLength           = 12
	PacketTotalLengthBytes = 4
)

var (
	ErrIllegalHeaderLength = errors.New("cmpp2 header length is invalid")
)

// Header CMPP PduCMPP 的公共 header
type Header struct {
	// 4 字节，消息总长度
	TotalLength uint32

	// 4 字节，命令或相应类型
	CommandID CommandID

	// 4 字节，消息流水号
	SequenceID uint32
}

func (h Header) Bytes() []byte {
	b := make([]byte, 0, HeaderLength)
	buf := bytes.NewBuffer(b)
	_ = binary.Write(buf, binary.BigEndian, h.TotalLength)
	_ = binary.Write(buf, binary.BigEndian, h.CommandID)
	_ = binary.Write(buf, binary.BigEndian, h.SequenceID)
	return buf.Bytes()
}

func NewHeader(totalLength uint32, commandID CommandID, sequenceID uint32) Header {
	return Header{TotalLength: totalLength, CommandID: commandID, SequenceID: sequenceID}
}

func NewHeaderFromBytes(d []byte) (h Header, err error) {
	if len(d) < MinCMPPPduLength {
		return h, ErrIllegalHeaderLength
	}

	buf := bytes.NewBuffer(d)
	return NewHeaderFromReader(buf)
}

func NewHeaderFromReader(buf io.Reader) (Header, error) {
	var h Header
	if err := binary.Read(buf, binary.BigEndian, &h.TotalLength); err != nil {
		return h, err
	}
	if err := binary.Read(buf, binary.BigEndian, &h.CommandID); err != nil {
		return h, err
	}
	if err := binary.Read(buf, binary.BigEndian, &h.SequenceID); err != nil {
		return h, err
	}
	return h, nil
}

// PeekHeader 尝试读取前 HeaderLength 长度的字节并解析成 Header， 不影响原有 reader 的游标
func PeekHeader(buf []byte) (h Header, err error) {
	if len(buf) < HeaderLength {
		return h, ErrIllegalHeaderLength
	}
	h.TotalLength = binary.BigEndian.Uint32(buf[:4])
	h.CommandID = CommandID(binary.BigEndian.Uint32(buf[4:8]))
	h.SequenceID = binary.BigEndian.Uint32(buf[8:12])
	return h, nil
}
