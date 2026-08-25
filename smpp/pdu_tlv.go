package smpp

import (
	"encoding/binary"
	"fmt"
	"math"
	"sort"

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
	b, err := t.MarshalBinary()
	if err != nil {
		return nil
	}
	return b
}

func (t TLV) MarshalBinary() ([]byte, error) {
	if len(t.ValueBytes) > math.MaxUint16 {
		return nil, fmt.Errorf("TLV %#04x value is %d octets; maximum is %d", t.Tag, len(t.ValueBytes), math.MaxUint16)
	}
	if int(t.Length) != len(t.ValueBytes) {
		return nil, fmt.Errorf("TLV %#04x length=%d does not match value length=%d", t.Tag, t.Length, len(t.ValueBytes))
	}

	b := make([]byte, 4+len(t.ValueBytes))
	binary.BigEndian.PutUint16(b[0:2], t.Tag)
	binary.BigEndian.PutUint16(b[2:4], t.Length)
	copy(b[4:], t.ValueBytes)
	return b, nil
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

type TLVList []TLV

func (t TLVList) MarshalBinary() ([]byte, error) {
	var total int
	for _, tlv := range t {
		if len(tlv.ValueBytes) > math.MaxUint16 || total > math.MaxInt-4-len(tlv.ValueBytes) {
			return nil, fmt.Errorf("TLV %#04x value is too large", tlv.Tag)
		}
		total += 4 + len(tlv.ValueBytes)
	}

	b := make([]byte, 0, total)
	for _, tlv := range t {
		encoded, err := tlv.MarshalBinary()
		if err != nil {
			return nil, err
		}
		b = append(b, encoded...)
	}
	return b, nil
}

func (t TLVList) Bytes() []byte {
	b, err := t.MarshalBinary()
	if err != nil {
		return nil
	}
	return b
}

func (t TLVList) All(tag uint16) []TLV {
	result := make([]TLV, 0)
	for _, tlv := range t {
		if tlv.Tag == tag {
			result = append(result, tlv)
		}
	}
	return result
}

func (t TLVList) ToMap() TLVs {
	if len(t) == 0 {
		return nil
	}
	result := make(TLVs, len(t))
	for _, tlv := range t {
		result[tlv.Tag] = tlv
	}
	return result
}

func (t TLVList) String() string {
	if len(t) == 0 {
		return ""
	}
	s := "\n"
	for _, tlv := range t {
		if !tlv.IsEmpty() {
			s += fmt.Sprintf("\t%s\n", tlv.String())
		}
	}
	return s
}

// ValidateTLVList validates structural and cross-field rules that are common
// to SMPP 3.4 optional parameters. PDU-specific allow-lists and conflicts are
// validated by the concrete PDU implementation.
func ValidateTLVList(tlvs TLVList) error {
	for _, tlv := range tlvs {
		if len(tlv.ValueBytes) > math.MaxUint16 || int(tlv.Length) != len(tlv.ValueBytes) {
			return fmt.Errorf("TLV %#04x length=%d does not match value length=%d", tlv.Tag, tlv.Length, len(tlv.ValueBytes))
		}
	}

	if values := tlvs.All(RECEIPTED_MESSAGE_ID); len(values) > 1 {
		return fmt.Errorf("receipted_message_id occurs %d times; maximum is one", len(values))
	} else if len(values) == 1 {
		value := values[0].ValueBytes
		if len(value) < 1 || len(value) > 65 || value[len(value)-1] != 0 {
			return invalidTLVValue(RECEIPTED_MESSAGE_ID, "must be a NULL-terminated C-Octet String of 1..65 octets")
		}
		if err := ValidateCString("receipted_message_id", string(value[:len(value)-1]), 65); err != nil {
			return err
		}
	}

	if values := tlvs.All(SC_INTERFACE_VERSION); len(values) > 1 {
		return fmt.Errorf("sc_interface_version occurs %d times; maximum is one", len(values))
	} else if len(values) == 1 {
		if len(values[0].ValueBytes) != 1 {
			return invalidTLVValue(SC_INTERFACE_VERSION, "length must be one")
		}
		if err := ValidateInterfaceVersion(values[0].ValueBytes[0]); err != nil {
			return err
		}
	}

	if values := tlvs.All(MESSAGE_PAYLOAD); len(values) > 1 {
		return fmt.Errorf("message_payload occurs %d times; maximum is one", len(values))
	}

	refs := tlvs.All(SAR_MSG_REF_NUM)
	totals := tlvs.All(SAR_TOTAL_SEGMENTS)
	sequences := tlvs.All(SAR_SEGMENT_SEQNUM)
	if len(refs)+len(totals)+len(sequences) > 0 {
		if len(refs) != 1 || len(totals) != 1 || len(sequences) != 1 {
			return fmt.Errorf("SAR optional parameters must contain exactly one reference, total, and sequence value")
		}
		if len(refs[0].ValueBytes) != 2 || len(totals[0].ValueBytes) != 1 || len(sequences[0].ValueBytes) != 1 {
			return invalidTLVValue(SAR_MSG_REF_NUM, "SAR reference/total/sequence lengths must be 2/1/1")
		}
		if totals[0].ValueBytes[0] == 0 {
			return invalidTLVValue(SAR_TOTAL_SEGMENTS, "must be greater than zero")
		}
		if sequences[0].ValueBytes[0] == 0 || sequences[0].ValueBytes[0] > totals[0].ValueBytes[0] {
			return invalidTLVValue(SAR_SEGMENT_SEQNUM, "must be in 1..sar_total_segments")
		}
	}

	return nil
}

func invalidTLVValue(tag uint16, detail string) error {
	return fmt.Errorf("TLV %#04x %s", tag, detail)
}

func ReadTLVList(r *packet.Reader) (TLVList, error) {
	if r.Remaining() == 0 {
		return nil, nil
	}
	if r.Error() != nil {
		return nil, r.Error()
	}

	tlvs := make(TLVList, 0)
	for r.Remaining() > 0 {
		if r.Remaining() < 4 {
			return nil, fmt.Errorf("truncated TLV header: %d octets remain", r.Remaining())
		}

		tag := r.ReadUint16()
		length := r.ReadUint16()
		if err := r.Error(); err != nil {
			return nil, err
		}
		if r.Remaining() < int(length) {
			return nil, fmt.Errorf("truncated TLV %#04x value: declared %d octets, %d remain", tag, length, r.Remaining())
		}
		value := r.ReadNBytes(int(length))
		if err := r.Error(); err != nil {
			return nil, err
		}

		tlvs = append(tlvs, TLV{
			Tag:        tag,
			Length:     length,
			ValueBytes: value,
		})
	}

	return tlvs, nil
}

func ReadTLVs(r *packet.Reader) (TLVs, error) {
	tlvs, err := ReadTLVList(r)
	if err != nil {
		return nil, err
	}
	return tlvs.ToMap(), nil
}

// ReadTLVs1 is retained for source compatibility. New decoders should use
// ReadTLVList or ReadTLVs so malformed optional parameter streams are reported.
func ReadTLVs1(r *packet.Reader) TLVs {
	tlvs, err := ReadTLVs(r)
	if err != nil {
		return nil
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
	b, err := t.MarshalBinary()
	if err != nil {
		return nil
	}
	return b
}

func (t TLVs) MarshalBinary() ([]byte, error) {
	if len(t) == 0 {
		return nil, nil
	}
	tags := make([]int, 0, len(t))
	for tag := range t {
		tags = append(tags, int(tag))
	}
	sort.Ints(tags)

	list := make(TLVList, 0, len(tags))
	for _, tag := range tags {
		list = append(list, t[uint16(tag)])
	}
	return list.MarshalBinary()
}

func (t TLVs) ToList() TLVList {
	if len(t) == 0 {
		return nil
	}
	tags := make([]int, 0, len(t))
	for tag := range t {
		tags = append(tags, int(tag))
	}
	sort.Ints(tags)
	result := make(TLVList, 0, len(tags))
	for _, tag := range tags {
		result = append(result, t[uint16(tag)])
	}
	return result
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
