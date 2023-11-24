package packet

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/valyala/bytebufferpool"
)

var packetOrder = binary.BigEndian

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

	return
}

func (p2 *Writer) WriteUint16(p uint16) {
	p2.writeNumeric(p)
	p2.written += 2

	return
}

func (p2 *Writer) WriteUint32(p uint32) {
	p2.writeNumeric(p)
	p2.written += 4

	return
}

func (p2 *Writer) WriteUint64(p uint64) {
	p2.writeNumeric(p)
	p2.written += 8

	return
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

// ----------------------------

type Reader struct {
	buffer  *bytes.Buffer
	opError *packetOptError
}

func NewPacketReader(data []byte) *Reader {
	return &Reader{buffer: bytes.NewBuffer(data)}
}

func (p *Reader) readNumeric(data interface{}) {
	if p.opError != nil {
		return
	}

	if err := binary.Read(p.buffer, packetOrder, data); err != nil {
		p.opError = newPacketError(err, "ReadNumeric")
	}
}

func (p *Reader) ReadUint8() uint8 {
	var res uint8
	p.readNumeric(&res)
	return res
}

func (p *Reader) ReadUint16() uint16 {
	var res uint16
	p.readNumeric(&res)
	return res
}

func (p *Reader) ReadUint32() uint32 {
	var res uint32
	p.readNumeric(&res)
	return res
}

func (p *Reader) ReadUint64() uint64 {
	var res uint64
	p.readNumeric(&res)
	return res
}

func (p *Reader) ReadBytes(receiver []byte) {
	if p.opError != nil {
		return
	}

	if len(receiver) == 0 {
		return
	}

	n, err := p.buffer.Read(receiver)
	if err != nil {
		p.opError = newPacketError(err, "ReadBytes read")
		return
	}

	if n < len(receiver) {
		p.opError = newPacketError(fmt.Errorf("the data read is less than expected"), "ReadBytes")
		return
	}
}

func (p *Reader) ReadCStringN(n int) string {
	if p.opError != nil {
		return ""
	}

	if n <= 0 {
		return ""
	}

	temp := make([]byte, n)

	r, err := p.buffer.Read(temp)
	if err != nil {
		p.opError = newPacketError(err, "ReadCStringN read")
		return ""
	}

	if r != n {
		p.opError = newPacketError(fmt.Errorf("read unexpected length"), "ReadBytes")
		return ""
	}

	if idx := bytes.IndexByte(temp, 0); idx > -1 {
		temp = temp[:idx]
	}

	return string(temp)
}

func (p *Reader) ReadCString() string {
	if p.opError != nil {
		return ""
	}

	line, err := p.buffer.ReadString(0x00)
	if err != nil {
		p.opError = newPacketError(err, "ReadCString read")
		return ""
	}
	if len(line) == 0 {
		return ""
	}

	return line[:len(line)-1]
}

func (p *Reader) ReadNBytes(n int) []byte {
	if p.opError != nil {
		return nil
	}

	if n <= 0 {
		return nil
	}

	temp := make([]byte, n)

	r, err := p.buffer.Read(temp)
	if err != nil {
		p.opError = newPacketError(err, "ReadCStringN read")
		return nil
	}

	if r != n {
		p.opError = newPacketError(fmt.Errorf("read unexpected length"), "ReadBytes")
		return nil
	}

	return temp
}

func (p *Reader) HexString() string {
	if p.opError != nil {
		return ""
	}

	return hex.EncodeToString(p.buffer.Bytes())
}

func (p *Reader) Error() error {
	if p.opError != nil {
		return p.opError
	}

	return nil
}

func (p *Reader) SetErrNil() {
	p.opError = nil
}

func (p *Reader) Release() {
	p.buffer.Reset()
	p.opError = nil
}

func (p *Reader) Remaining() int {
	return p.buffer.Len()
}

// ----------------------------

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
