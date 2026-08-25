package packet

import (
	"encoding/binary"
	"errors"
	"fmt"
)

var packetOrder = binary.BigEndian

// ErrFieldLengthExceeded indicates that a field value is longer than the
// limit accepted by a packet writer operation.
var ErrFieldLengthExceeded = errors.New("field length exceeds limit")

// FieldLengthError reports a field value whose byte length exceeds its limit.
// Actual and Limit are measured in bytes (octets), not runes.
type FieldLengthError struct {
	Field  string
	Actual int
	Limit  int
}

// Error returns a safe description without including the field value.
func (e *FieldLengthError) Error() string {
	return fmt.Sprintf("field %q actual length %d exceeds limit %d (excess %d)", e.Field, e.Actual, e.Limit, e.Actual-e.Limit)
}

// Is reports whether the error is a field-length overflow.
func (e *FieldLengthError) Is(target error) bool {
	return target == ErrFieldLengthExceeded
}

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
