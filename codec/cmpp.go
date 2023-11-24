package codec

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/hujm2023/go-sms-protocol/cmpp"
)

type CMPPCodec struct{}

func NewCMPPCodec() *CMPPCodec {
	return new(CMPPCodec)
}

// Decode implements sticky packet handling for the CMPP protocol,
// reading a complete CMPP packet from the ConnReader.
// If the data is incomplete, it will return ErrPacketNotComplete.
func (cc *CMPPCodec) Decode(c ConnReader) ([]byte, error) {
	// peek the `length` field
	totalLenBytes, _ := c.Peek(cmpp.PacketTotalLengthBytes)
	if len(totalLenBytes) < cmpp.PacketTotalLengthBytes {
		return nil, ErrPacketNotComplete
	}

	totalLen := int(binary.BigEndian.Uint32(totalLenBytes))
	if c.Size() < totalLen {
		return nil, ErrPacketNotComplete
	}

	// peek the left by `totalLen`
	buf, _ := c.Peek(totalLen)
	if len(buf) < totalLen {
		return nil, ErrPacketNotComplete
	}

	// update cursor, make `read` done
	n, err := c.Discard(totalLen)
	if err != nil {
		return nil, fmt.Errorf("discard total packet error: %w", err)
	}
	if n != len(buf) {
		return nil, fmt.Errorf("not discard enough data")
	}

	return buf, nil
}

func (cc *CMPPCodec) DecodeBlocked(c ConnReader) ([]byte, error) {
	totalLenBytes := make([]byte, cmpp.PacketTotalLengthBytes)
	_, err := io.ReadFull(c, totalLenBytes)
	if err != nil {
		return nil, err
	}
	totalLen := int(binary.BigEndian.Uint32(totalLenBytes))

	left := make([]byte, totalLen)
	_, err = io.ReadFull(c, left[cmpp.PacketTotalLengthBytes:])
	if err != nil {
		return nil, err
	}

	copy(left[:cmpp.PacketTotalLengthBytes], totalLenBytes)

	return left, nil
}
