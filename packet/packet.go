package packet

import (
	"encoding/binary"
)

var packetOrder = binary.BigEndian

type packetOptError struct {
	err error
	op  string
}

func newPacketError(e error, op string) *packetOptError {
	return &packetOptError{
		err: e,
		op:  op,
	}
}

func (e *packetOptError) Error() string {
	if e.err == nil {
		return "<nil>"
	}
	return e.op + " error: " + e.err.Error()
}

func (e *packetOptError) Cause() error {
	return e.err
}

func (e *packetOptError) Unwrap() error {
	return e.err
}

func (e *packetOptError) Op() string {
	return e.op
}
