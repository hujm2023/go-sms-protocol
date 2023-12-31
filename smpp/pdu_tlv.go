package smpp

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/hujm2023/go-sms-protocol/packet"
)

type TLV struct {
	Tag    uint16
	Length uint16
	Value  []byte
}

func NewTLV(tag uint16, value []byte) TLV {
	return TLV{
		Tag:    tag,
		Length: uint16(len(value)),
		Value:  value,
	}
}

func NewTLVS(tag uint16, value string) TLV {
	return NewTLV(tag, []byte(value))
}

func (t *TLV) Bytes() []byte {
	b := make([]byte, t.Length+4)
	binary.BigEndian.PutUint16(b[0:2], t.Tag)
	binary.BigEndian.PutUint16(b[2:4], t.Length)
	copy(b[4:], t.Value)
	return b
}

func ReadTLVs(r *packet.Reader) (map[uint16]TLV, error) {
	if r.Remaining() == 0 {
		return nil, nil
	}

	if r.Error() != nil {
		return nil, nil
	}

	tlvs := make(map[uint16]TLV)
	for {
		if r.Remaining() == 0 {
			return tlvs, nil
		}

		temp := make([]byte, 4)
		r.ReadBytes(temp)
		if e := r.Error(); e != nil {
			if errors.Is(e, io.EOF) {
				r.SetErrNil()
				break
			}
			return nil, e
		}

		tag := binary.BigEndian.Uint16(temp[:2])
		length := binary.BigEndian.Uint16(temp[2:4])

		// read left value
		value := make([]byte, length)
		r.ReadBytes(value)
		if e := r.Error(); e != nil {
			if errors.Is(r.Error(), io.EOF) {
				r.SetErrNil()
				break
			}
			return nil, e
		}

		tlvs[tag] = TLV{
			Tag:    tag,
			Length: length,
			Value:  value,
		}
	}

	return tlvs, nil
}

func ReadTLVs1(r *packet.Reader) TLVs {
	if r.Remaining() == 0 {
		return nil
	}

	if r.Error() != nil {
		return nil
	}

	tlvs := make(map[uint16]TLV)
	for {
		if r.Remaining() == 0 {
			return tlvs
		}

		temp := make([]byte, 4)
		r.ReadBytes(temp)
		if e := r.Error(); e != nil {
			if errors.Is(e, io.EOF) {
				r.SetErrNil()
				break
			}
			return nil
		}

		tag := binary.BigEndian.Uint16(temp[:2])
		length := binary.BigEndian.Uint16(temp[2:4])

		// read left value
		value := make([]byte, length)
		r.ReadBytes(value)
		if e := r.Error(); e != nil {
			if errors.Is(r.Error(), io.EOF) {
				r.SetErrNil()
				break
			}
			return nil
		}

		tlvs[tag] = TLV{
			Tag:    tag,
			Length: length,
			Value:  value,
		}
	}

	return tlvs
}

type TLVs map[uint16]TLV

func (t TLVs) Bytes() []byte {
	b := make([]byte, 0)
	for _, tlv := range t {
		b = append(b, tlv.Bytes()...)
	}
	return b
}
