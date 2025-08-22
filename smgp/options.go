package smgp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	"github.com/hujm2023/go-sms-protocol/packet"
)

var ErrLength = errors.New("Options: error length")

type Tag uint16

// 可选参数标签定义  Option Tag
const (
	TAG_TP_pid Tag = 0x0001 + iota
	TAG_TP_udhi
	TAG_LinkID
	TAG_ChargeUserType
	TAG_ChargeTermType
	TAG_ChargeTermPseudo
	TAG_DestTermType
	TAG_DestTermPseudo
	TAG_PkTotal
	TAG_PkNumber
	TAG_SubmitMsgType
	TAG_SPDealResult
	TAG_SrcTermType
	TAG_SrcTermPseudo
	TAG_NodesCount
	TAG_MsgSrc
	TAG_SrcType
	TAG_MServiceID
)

type Option struct {
	tag    uint16
	length uint16
	value  []byte
}

func NewOption(tag Tag, value []byte) Option {
	return Option{
		tag:    uint16(tag),
		length: uint16(len(value)),
		value:  value,
	}
}

func (o Option) String() string {
	return fmt.Sprintf("Option{tag:%d, length:%d, value:%v}", o.tag, o.length, o.value)
}

func (o Option) Len() int {
	return int(o.length)
}

func (o Option) Value() []byte {
	return o.value
}

func (o Option) Bytes() []byte {
	b := make([]byte, o.length+4)
	binary.BigEndian.PutUint16(b[:2], o.tag)
	binary.BigEndian.PutUint16(b[2:4], o.length)
	copy(b[4:], o.value)
	return b
}

func (o Option) IsEmpty() bool {
	return o.tag == 0 && o.length == 0 && len(o.value) == 0
}

// ---------------------------------------------------------------------------------------

// 可选参数 map
type Options map[Tag]Option

func (o Options) Add(opt Option) {
	if o == nil {
		o = make(Options)
	}

	o[Tag(opt.tag)] = opt
}

func (o Options) String() string {
	if len(o) == 0 {
		return ""
	}
	s := "\n"
	for idx := range o {
		if o[idx].IsEmpty() {
			continue
		}
		s += fmt.Sprintf("\t%s\n", o[idx].String())
	}

	return s
}

// 返回可选字段部分的长度
func (o Options) Len() int {
	length := 0

	for _, v := range o {
		length += 2 + 2 + len(v.value)
	}

	return length
}

func (o Options) Serialize() []byte {
	b := make([]byte, 0)
	for _, v := range o {
		b = append(b, v.Bytes()...)
	}
	return b
}

func (o Options) TP_udhi() uint8 {
	if val, exist := o[TAG_TP_udhi]; exist {
		return val.value[0]
	}
	return 0
}

func ParseOptions(rawData []byte) (Options, error) {
	var (
		p      = 0
		ops    = make(Options)
		length = len(rawData)
	)

	for p < length {
		if length-p < 2+2 { // less than Tag len + Length len
			return nil, ErrLength
		}

		tag := binary.BigEndian.Uint16(rawData[p:])
		p += 2

		vlen := binary.BigEndian.Uint16(rawData[p:])
		p += 2

		if length-p < int(vlen) { // remaining not enough
			return nil, ErrLength
		}

		value := rawData[p : p+int(vlen)]
		p += int(vlen)

		ops[Tag(tag)] = Option{
			tag:    tag,
			length: vlen,
			value:  value,
		}
	}

	return ops, nil
}

func ReadOptions(r *packet.Reader) Options {
	if r.Remaining() == 0 {
		return nil
	}

	if r.Error() != nil {
		return nil
	}

	options := make(Options)
	for {
		if r.Remaining() == 0 {
			return options
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

		options[Tag(tag)] = Option{
			tag:    tag,
			length: length,
			value:  value,
		}
	}

	return options
}
