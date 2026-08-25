package packet

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/valyala/bytebufferpool"
)

type Writer struct {
	buf     *bytebufferpool.ByteBuffer
	written int
	opError *packetOptError
}

func NewPacketWriter(totalLen ...int) *Writer {
	return &Writer{buf: bytebufferpool.Get()}
}

func (p2 *Writer) writeNumeric(p any) bool {
	if p2.opError != nil {
		return false
	}

	if err := binary.Write(p2.buf, packetOrder, p); err != nil {
		p2.opError = newPacketError(err, "WriteNumeric write")
		return false
	}
	return true
}

func (p2 *Writer) WriteUint8(p uint8) {
	if p2.writeNumeric(p) {
		p2.written += 1
	}
}

func (p2 *Writer) WriteUint16(p uint16) {
	if p2.writeNumeric(p) {
		p2.written += 2
	}
}

func (p2 *Writer) WriteUint32(p uint32) {
	if p2.writeNumeric(p) {
		p2.written += 4
	}
}

func (p2 *Writer) WriteUint64(p uint64) {
	if p2.writeNumeric(p) {
		p2.written += 8
	}
}

func (p2 *Writer) WriteBytes(data []byte) {
	if p2.opError != nil {
		return
	}

	n, err := p2.buf.Write(data)
	if err != nil {
		p2.opError = newPacketError(err, "WriteBytes write")
		return
	}
	if n != len(data) {
		p2.opError = newPacketError(io.ErrShortWrite, "WriteBytes write")
		return
	}

	p2.written += n
}

func (p2 *Writer) WriteString(s string) {
	if p2.opError != nil {
		return
	}

	n, err := p2.buf.WriteString(s)
	if err != nil {
		p2.opError = newPacketError(err, "WriteString write")
		return
	}

	if n != len(s) {
		p2.opError = newPacketError(io.ErrShortWrite, "WriteString not finished")
		return
	}

	p2.written += n
}

func (p2 *Writer) WriteCString(s string) {
	if p2.opError != nil {
		return
	}

	var n int
	var err error
	if len(s) > 0 {
		n, err = p2.buf.WriteString(s)
		if err != nil {
			p2.opError = newPacketError(err, "WriteString write")
			return
		}
	}

	err = p2.buf.WriteByte(0x00)
	if err != nil {
		p2.opError = newPacketError(err, "WriteString write")
		return
	}

	if n != len(s) {
		p2.opError = newPacketError(io.ErrShortWrite, "WriteString not finished")
		return
	}

	p2.written += n + 1
}

func (p2 *Writer) WriteFixedLenString(s string, n int) {
	p2.writeFixedLenString("", s, n, "WriteFixedLenString write")
}

// WriteFixedLenStringField writes a zero-padded fixed-width string and
// reports an overflow with the supplied protocol field name.
func (p2 *Writer) WriteFixedLenStringField(field, value string, limit int) {
	p2.writeFixedLenString(field, value, limit, "WriteFixedLenStringField write")
}

func (p2 *Writer) writeFixedLenString(field, value string, limit int, op string) {
	if p2.opError != nil {
		return
	}

	if len(value) > limit {
		p2.opError = newPacketError(&FieldLengthError{
			Field:  field,
			Actual: len(value),
			Limit:  limit,
		}, op)
		return
	}

	nn, err := p2.buf.WriteString(strings.Join([]string{value, string(make([]byte, limit-len(value)))}, ""))
	if err != nil {
		p2.opError = newPacketError(err, op)
		return
	}

	if nn != limit {
		p2.opError = newPacketError(fmt.Errorf("unexpected written bytes"), op)
		return
	}

	p2.written += nn
}

func (p2 *Writer) Bytes() (data []byte, err error) {
	if p2.opError != nil {
		return nil, p2.opError
	}

	res := make([]byte, p2.buf.Len())
	copy(res, p2.buf.Bytes())

	return res, nil
}

func (p2 *Writer) BytesWithLength() (data []byte, err error) {
	if p2.opError != nil {
		return nil, p2.opError
	}

	res := make([]byte, 4+p2.written)
	packetOrder.PutUint32(res, uint32(p2.written)+4)
	copy(res[4:], p2.buf.Bytes())

	return res, nil
}

func (p2 *Writer) Written() int {
	return p2.written
}

func (p2 *Writer) HexString() string {
	if p2.opError != nil {
		return ""
	}
	return hex.EncodeToString(p2.buf.Bytes())
}

func (p2 *Writer) Error() error {
	if p2.opError != nil {
		return p2.opError
	}

	return nil
}

func (p2 *Writer) Release() {
	bytebufferpool.Put(p2.buf)
	p2.written = 0
	p2.opError = nil
}

func (p2 *Writer) Len() int {
	if p2.opError != nil {
		return 0
	}
	return p2.buf.Len()
}
