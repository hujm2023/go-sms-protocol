package cmpp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/hujm2023/go-sms-protocol/packet"
)

const (
	// HeaderLength defines the fixed length of the CMPP PDU header.
	HeaderLength = 12
	// PacketTotalLengthBytes defines the number of bytes used for the TotalLength field in the header.
	PacketTotalLengthBytes = 4
)

var (
	// ErrIllegalHeaderLength indicates that the provided data length is less than the required header length.
	ErrIllegalHeaderLength = errors.New("cmpp2 header length is invalid")
	// ErrInvalidPudLength indicates that the PDU length specified in the header is invalid or inconsistent.
	ErrInvalidPudLength = errors.New("invalid pdu length")
)

// Header represents the common header structure for all CMPP PDUs.
type Header struct {
	// TotalLength is the total length of the PDU in bytes, including the header.
	TotalLength uint32

	// CommandID identifies the type of the CMPP command or response.
	CommandID CommandID

	// SequenceID is the sequence number of the message, used for matching requests and responses.
	SequenceID uint32
}

// Bytes encodes the Header struct into a byte slice.
func (h Header) Bytes() []byte {
	b := make([]byte, 0, HeaderLength)
	buf := bytes.NewBuffer(b)
	_ = binary.Write(buf, binary.BigEndian, h.TotalLength)
	_ = binary.Write(buf, binary.BigEndian, h.CommandID)
	_ = binary.Write(buf, binary.BigEndian, h.SequenceID)
	return buf.Bytes()
}

// String returns a string representation of the Header.
func (h Header) String() string {
	return fmt.Sprintf("{TotalLength:%d, CommandID:%s, SequenceID:%d}", h.TotalLength, h.CommandID.String(), h.SequenceID)
}

// NewHeader creates a new Header instance.
func NewHeader(totalLength uint32, commandID CommandID, sequenceID uint32) Header {
	return Header{TotalLength: totalLength, CommandID: commandID, SequenceID: sequenceID}
}

// NewHeaderFromBytes decodes a Header from a byte slice.
func NewHeaderFromBytes(d []byte) (h Header, err error) {
	if len(d) < MinCMPPPduLength {
		return h, ErrIllegalHeaderLength
	}

	buf := bytes.NewBuffer(d)
	return NewHeaderFromReader(buf)
}

// NewHeaderFromReader decodes a Header from an io.Reader.
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

// PeekHeader reads the header bytes from a byte slice without consuming them
// and decodes them into a Header struct.
func PeekHeader(buf []byte) (h Header, err error) {
	if len(buf) < HeaderLength {
		return h, ErrIllegalHeaderLength
	}
	h.TotalLength = binary.BigEndian.Uint32(buf[:4])
	h.CommandID = CommandID(binary.BigEndian.Uint32(buf[4:8]))
	h.SequenceID = binary.BigEndian.Uint32(buf[8:12])
	return h, nil
}

// ReadHeader reads and decodes a Header from a packet.Reader.
func ReadHeader(r *packet.Reader) Header {
	var h Header
	h.TotalLength = r.ReadUint32()
	h.CommandID = CommandID(r.ReadUint32())
	h.SequenceID = r.ReadUint32()
	return h
}

// WriteHeader encodes and writes a Header to a packet.Writer, including the TotalLength.
func WriteHeader(h Header, buf *packet.Writer) {
	buf.WriteUint32(h.TotalLength)
	buf.WriteUint32(uint32(h.CommandID))
	buf.WriteUint32(h.SequenceID)
}

// WriteHeaderNoLength encodes and writes a Header to a packet.Writer, excluding the TotalLength.
// This is typically used when the total length needs to be calculated and written later.
func WriteHeaderNoLength(h Header, buf *packet.Writer) {
	buf.WriteUint32(uint32(h.CommandID))
	buf.WriteUint32(h.SequenceID)
}
