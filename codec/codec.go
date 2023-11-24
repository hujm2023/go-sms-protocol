package codec

import (
	"errors"
	"io"
)

var ErrPacketNotComplete = errors.New("packet not completed")

// Codec handles TCP sticky packet problem
type Codec interface {
	// Decode tries to read a complete packet from the tcp connection without blocking
	// - If a complete packet is read, it returns nil;
	// - If the packet is incomplete, it returns ErrPacketNotComplete;
	// - For other errors, the connection should be closed
	Decode(c ConnReader) ([]byte, error)

	// DecodeBlocked blocks reading a packet
	DecodeBlocked(c ConnReader) ([]byte, error)
}

type ConnReader interface {
	io.Reader

	// Peek returns the next n bytes without advancing the reader. The bytes stop
	// being valid at the next read call. If Peek returns fewer than n bytes, it
	// also returns an error explaining why the read is short. The error is
	// ErrBufferFull if n is larger than b's buffer size.
	Peek(n int) ([]byte, error)

	// Discard skips the next n bytes, returning the number of bytes discarded.
	//
	// If Discard skips fewer than n bytes, it also returns an error.
	// If 0 <= n <= Conn.Size(), Discard is guaranteed to succeed without
	// reading from the underlying io.Reader.
	Discard(n int) (int, error)

	// Size returns the number of bytes that can be read from current Connection.
	Size() int

	// LocalAddress ...
	LocalAddress() string

	// RemoteAddress ...
	RemoteAddress() string
}
