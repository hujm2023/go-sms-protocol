package codec

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/hujm2023/go-sms-protocol/smpp"
)

// SMPPCodec provides methods for encoding and decoding SMPP PDUs.
// It handles sticky packets for the SMPP protocol.
type SMPPCodec struct{}

// NewSMPPCodec creates and returns a new SMPPCodec instance.
func NewSMPPCodec() *SMPPCodec {
	return new(SMPPCodec)
}

// Decode implements sticky packet handling for the SMPP protocol,
// reading a complete SMPP packet from the ConnReader.
// It attempts to read without blocking.
// If the data is incomplete, it will return ErrPacketNotComplete.
func (cc *SMPPCodec) Decode(c ConnReader) ([]byte, error) {
	totalLenBytes, _ := c.Peek(smpp.MinSMPPHeaderLen)
	if len(totalLenBytes) < smpp.MinSMPPHeaderLen {
		return nil, ErrPacketNotComplete
	}

	totalLen := int(binary.BigEndian.Uint32(totalLenBytes))
	if c.Size() < totalLen {
		return nil, ErrPacketNotComplete
	}

	buf, _ := c.Peek(totalLen)
	if len(buf) < totalLen {
		return nil, ErrPacketNotComplete
	}

	n, err := c.Discard(totalLen)
	if err != nil {
		return nil, fmt.Errorf("discard total packet error: %w", err)
	}
	if n != len(buf) {
		return nil, fmt.Errorf("not discard enough data")
	}

	return buf, nil
}

// DecodeBlocked reads a complete SMPP packet from the ConnReader in a blocking manner.
// It reads the command length first, then reads the remaining bytes.
func (cc *SMPPCodec) DecodeBlocked(c ConnReader) ([]byte, error) {
	totalLenBytes := make([]byte, smpp.MinSMPPHeaderLen)
	_, err := io.ReadFull(c, totalLenBytes)
	if err != nil {
		return nil, err
	}
	totalLen := int(binary.BigEndian.Uint32(totalLenBytes))

	left := make([]byte, totalLen)
	_, err = io.ReadFull(c, left[smpp.MinSMPPHeaderLen:])
	if err != nil {
		return nil, err
	}
	copy(left[:smpp.MinSMPPHeaderLen], totalLenBytes)

	return left, nil
}
