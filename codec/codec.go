package codec

import (
	"errors"
	"io"
)

// ErrPacketNotComplete indicates that the reader has not received a full packet yet.
var ErrPacketNotComplete = errors.New("packet not completed")

// Codec defines the interface for handling protocol-specific packet encoding and decoding,
// particularly addressing the TCP sticky packet problem.
type Codec interface {
	// Decode attempts to read a complete packet from the ConnReader without blocking.
	// It returns the byte slice of the complete packet if successful.
	// If the packet is incomplete, it returns ErrPacketNotComplete.
	// For other errors, the connection should typically be closed.
	Decode(c ConnReader) ([]byte, error)

	// DecodeBlocked reads a complete packet from the ConnReader in a blocking manner.
	// It waits until a full packet is available.
	DecodeBlocked(c ConnReader) ([]byte, error)
}

// ConnReader defines the interface for reading data from a connection,
// providing methods for peeking, discarding, and checking the size of readable data.
type ConnReader interface {
	io.Reader

	// Peek returns the next n bytes without advancing the reader.
	// The returned bytes are only valid until the next read call.
	// If fewer than n bytes are available, it returns an error.
	// Returns ErrBufferFull if n exceeds the buffer size.
	Peek(n int) ([]byte, error)

	// Discard skips the next n bytes, returning the number of bytes discarded.
	// If fewer than n bytes are skipped, it returns an error.
	// If 0 <= n <= ConnReader.Size(), Discard is guaranteed to succeed
	// without reading from the underlying io.Reader.
	Discard(n int) (int, error)

	// Size returns the number of bytes currently available for reading from the connection.
	Size() int
}
