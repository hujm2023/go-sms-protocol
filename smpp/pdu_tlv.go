package smpp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/hujm2023/go-sms-protocol/packet"
)

type TLV struct {
	Tag        uint16
	Length     uint16
	ValueBytes []byte
}

func NewTLV(tag uint16, value []byte) TLV {
	return TLV{
		Tag:        tag,
		Length:     uint16(len(value)),
		ValueBytes: value,
	}
}

func NewTLVByString(tag uint16, value string) TLV {
	return NewTLV(tag, []byte(value))
}

func (t TLV) Bytes() []byte {
	b := make([]byte, t.Length+4)
	binary.BigEndian.PutUint16(b[0:2], t.Tag)
	binary.BigEndian.PutUint16(b[2:4], t.Length)
	copy(b[4:], t.ValueBytes)
	return b
}

func (t TLV) Value() []byte {
	return t.ValueBytes
}

func (t TLV) IsEmpty() bool {
	return t.Tag == 0 && t.Length == 0 && len(t.ValueBytes) == 0
}

func (t TLV) String() string {
	if t.IsEmpty() {
		return ""
	}
	return fmt.Sprintf("TLV{Tag=%#x, Length=%d, Value=%v, ValueString=%s}", t.Tag, t.Length, t.ValueBytes, string(t.ValueBytes))
}

func ReadTLVs(r *packet.Reader) (TLVs, error) {
	if r.Remaining() == 0 {
		return nil, nil
	}

	if r.Error() != nil {
		return nil, r.Error()
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
			Tag:        tag,
			Length:     length,
			ValueBytes: value,
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
			Tag:        tag,
			Length:     length,
			ValueBytes: value,
		}
	}

	return tlvs
}

type TLVs map[uint16]TLV

func (t *TLVs) SetTLV(tlv TLV) {
	if *t == nil {
		*t = make(TLVs)
	}
	(*t)[tlv.Tag] = tlv
}

func (t TLVs) Bytes() []byte {
	b := make([]byte, 0)
	for _, tlv := range t {
		b = append(b, tlv.Bytes()...)
	}
	return b
}

func (t TLVs) String() string {
	if len(t) == 0 {
		return ""
	}
	s := "\n"
	for idx := range t {
		if t[idx].IsEmpty() {
			continue
		}
		s += fmt.Sprintf("\t%s\n", t[idx].String())
	}

	return s
}
