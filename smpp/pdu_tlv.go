package smpp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/hujm2023/go-sms-protocol/packet"
)

type TLV struct {
	tag    uint16
	length uint16
	value  []byte
}

func NewTLV(tag uint16, value []byte) TLV {
	return TLV{
		tag:    tag,
		length: uint16(len(value)),
		value:  value,
	}
}

func NewTLVByString(tag uint16, value string) TLV {
	return NewTLV(tag, []byte(value))
}

func (t TLV) Bytes() []byte {
	b := make([]byte, t.length+4)
	binary.BigEndian.PutUint16(b[0:2], t.tag)
	binary.BigEndian.PutUint16(b[2:4], t.length)
	copy(b[4:], t.value)
	return b
}

func (t TLV) Value() []byte {
	return t.value
}

func (t TLV) IsEmpty() bool {
	return t.tag == 0 && t.length == 0 && len(t.value) == 0
}

func (t TLV) String() string {
	if t.IsEmpty() {
		return ""
	}
	return fmt.Sprintf("TLV{Tag=%#x, Length=%d, Value=%v, ValueString=%s}", t.tag, t.length, t.value, string(t.value))
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
			tag:    tag,
			length: length,
			value:  value,
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
			tag:    tag,
			length: length,
			value:  value,
		}
	}

	return tlvs
}

type TLVs map[uint16]TLV

func (t *TLVs) SetTLV(tlv TLV) {
	if *t == nil {
		*t = make(TLVs)
	}
	(*t)[tlv.tag] = tlv
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
