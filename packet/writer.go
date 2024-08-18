package packet

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
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

func (p2 *Writer) writeNumeric(p any) {
	if p2.opError != nil {
		return
	}

	if err := binary.Write(p2.buf, packetOrder, p); err != nil {
		p2.opError = newPacketError(err, "WriteNumeric write")
		return
	}
}

func (p2 *Writer) WriteUint8(p uint8) {
	p2.writeNumeric(p)
	p2.written += 1
}

func (p2 *Writer) WriteUint16(p uint16) {
	p2.writeNumeric(p)
	p2.written += 2
}

func (p2 *Writer) WriteUint32(p uint32) {
	p2.writeNumeric(p)
	p2.written += 4
}

func (p2 *Writer) WriteUint64(p uint64) {
	p2.writeNumeric(p)
	p2.written += 8
}

func (p2 *Writer) WriteBytes(data []byte) {
	if p2.opError != nil {
		return
	}

	n, err := p2.buf.Write(data)
	if err != nil || n != len(data) {
		p2.opError = newPacketError(err, "WriteBytes write")
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
		p2.opError = newPacketError(err, "WriteString not finished")
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
		p2.opError = newPacketError(err, "WriteString not finished")
		return
	}

	p2.written += n + 1
}

func (p2 *Writer) WriteFixedLenString(s string, n int) {
	if p2.opError != nil {
		return
	}

	if len(s) > n {
		p2.opError = newPacketError(fmt.Errorf("s is longer than the defined length"), "WriteFixedLenString write")
		return
	}

	nn, err := p2.buf.WriteString(strings.Join([]string{s, string(make([]byte, n-len(s)))}, ""))
	if err != nil {
		p2.opError = newPacketError(err, "WriteFixedLenString write")
		return
	}

	if nn != n {
		p2.opError = newPacketError(fmt.Errorf("unexpected written bytes"), "WriteFixedLenString write")
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
