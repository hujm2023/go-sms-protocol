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

// Error returns the error message string.
func (e *packetOptError) Error() string {
	if e.err == nil {
		return "<nil>"
	}
	return e.op + " error: " + e.err.Error()
}

// Cause returns the underlying cause of the error.
func (e *packetOptError) Cause() error {
	return e.err
}

// Unwrap returns the underlying error.
func (e *packetOptError) Unwrap() error {
	return e.err
}

// Op returns the operation during which the error occurred.
func (e *packetOptError) Op() string {
	return e.op
}
