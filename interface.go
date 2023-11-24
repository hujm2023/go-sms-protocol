package protocol

import (
	"errors"
)

var (
	ErrPacketNotComplete = errors.New("packet not completed")
	ErrUnsupportedPacket = errors.New("unsupported packed")
)

// PDU stands for Protocol Data Unit, which is the package for standard SMS protocols.
type PDU interface {
	// IEncode serializes a PDU into binary.
	IEncode() ([]byte, error)

	// IDecode deserializes binary into a PDU.
	IDecode(data []byte) error

	// SetSequenceID sets sequenceID for a PDU.
	SetSequenceID(id uint32)
}
