// Package exampleutil contains transport and lifecycle helpers shared by examples.
// It is internal to examples and is not a public client/session API.
package exampleutil

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/cloudwego/netpoll"

	protocol "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/nioserver"
)

type Decoder func([]byte) (protocol.PDU, error)

func frameSize(prefix []byte, minimum, maximum uint32) (int, error) {
	size := binary.BigEndian.Uint32(prefix)
	if size < minimum || size > maximum {
		return 0, fmt.Errorf("invalid frame length: %d (allowed %d..%d)", size, minimum, maximum)
	}
	return int(size), nil
}

func Unpack(minimum, maximum uint32, decode Decoder) nioserver.UnpackFunc {
	return func(_ context.Context, reader netpoll.Reader) (pdu protocol.PDU, err error) {
		defer func() { err = errors.Join(err, reader.Release()) }()
		prefix, err := reader.Peek(4)
		if err != nil {
			return nil, err
		}
		size, err := frameSize(prefix, minimum, maximum)
		if err != nil {
			return nil, err
		}
		frame, err := reader.ReadBinary(size)
		if err != nil {
			return nil, err
		}
		return decodeFrame(frame, decode)
	}
}

func Read(reader io.Reader, minimum, maximum uint32, decode Decoder) (protocol.PDU, error) {
	var prefix [4]byte
	if _, err := io.ReadFull(reader, prefix[:]); err != nil {
		return nil, err
	}
	size, err := frameSize(prefix[:], minimum, maximum)
	if err != nil {
		return nil, err
	}
	frame := make([]byte, size)
	copy(frame, prefix[:])
	if _, err := io.ReadFull(reader, frame[4:]); err != nil {
		return nil, err
	}
	return decodeFrame(frame, decode)
}

func decodeFrame(frame []byte, decode Decoder) (protocol.PDU, error) {
	pdu, err := decode(frame)
	if err != nil {
		return nil, err
	}
	canonical, err := pdu.IEncode()
	if err != nil {
		return nil, err
	}
	if len(canonical) != len(frame) {
		return nil, fmt.Errorf("PDU structure uses %d bytes, frame contains %d", len(canonical), len(frame))
	}
	return pdu, nil
}

func Write(writer io.Writer, pdu protocol.PDU) error {
	data, err := pdu.IEncode()
	if err != nil {
		return err
	}
	for len(data) > 0 {
		n, err := writer.Write(data)
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
		data = data[n:]
	}
	return nil
}

// Dial interrupts reads and writes when ctx is cancelled, including on Go 1.20.
// The caller must call cleanup to close the socket and join the watcher.
func Dial(ctx context.Context, address string) (net.Conn, func(), error) {
	var dialer net.Dialer
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return nil, nil, err
	}
	if deadline, ok := ctx.Deadline(); ok {
		if err := conn.SetDeadline(deadline); err != nil {
			_ = conn.Close()
			return nil, nil, err
		}
	}
	done := make(chan struct{})
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		select {
		case <-ctx.Done():
			_ = conn.Close()
		case <-done:
		}
	}()
	return conn, func() { close(done); _ = conn.Close(); <-stopped }, nil
}
