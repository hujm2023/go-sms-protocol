package packet

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

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

func (p *Reader) Bytes() []byte {
	if p.opError != nil {
		return []byte{}
	}
	return p.buffer.Bytes()
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

func (p *Reader) ReadCStringNWithoutTrim(n int) string {
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
