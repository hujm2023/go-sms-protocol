package smpp

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/hujm2023/go-sms-protocol/packet"
)

type TLVTestSuite struct {
	suite.Suite

	tag    uint16
	length uint16
	value  []byte

	valueBytes []byte
}

func (s *TLVTestSuite) SetupTest() {
	s.tag = RECEIPTED_MESSAGE_ID
	s.length = 6
	s.value = []byte("123456")
	s.valueBytes = []byte{
		0, 0x1e,
		0, 6,
		49, 50, 51, 52, 53, 54,
	}
}

func (s *TLVTestSuite) TestTLV_Bytes() {
	tt := NewTLV(s.tag, s.value)
	assert.True(s.T(), bytes.Equal(tt.Bytes(), s.valueBytes))
}

func (s *TLVTestSuite) TestReadTLVs() {
	r := packet.NewPacketReader(s.valueBytes)
	defer r.Release()

	t, err := ReadTLVs(r)
	assert.Nil(s.T(), err)
	assert.Equal(s.T(), 1, len(t))
	tlv := t[s.tag]
	assert.Equal(s.T(), s.tag, tlv.Tag)
	assert.Equal(s.T(), s.length, tlv.Length)
	assert.True(s.T(), bytes.Equal(tlv.ValueBytes, s.value))
}

func (s *TLVTestSuite) TestTLVString() {
	r := packet.NewPacketReader(s.valueBytes)
	defer r.Release()
	t, err := ReadTLVs(r)
	assert.Nil(s.T(), err)
	tlv := t[s.tag]

	s.T().Log(tlv.String())
	assert.Equal(s.T(), "TLV{Tag=0x1e, Length=6, Value=[49 50 51 52 53 54], ValueString=123456}", tlv.String())

	t[0x20] = TLV{
		Tag:        0x20,
		Length:     2,
		ValueBytes: []byte{49, 50},
	}
	s.T().Log(t.String())
}

func TestTLV(t *testing.T) {
	suite.Run(t, new(TLVTestSuite))
}

func TestStringerForTLVS(t *testing.T) {
	tlvs := TLVs(make(map[uint16]TLV))
	tlvs.SetTLV(TLV{
		Tag:        0x1e,
		Length:     6,
		ValueBytes: []byte("123456"),
	})
	tlvs.SetTLV(TLV{
		Tag:        0x20,
		Length:     2,
		ValueBytes: []byte{49, 50},
	})

	s := packet.NewPDUStringer()
	defer s.Release()

	s.OmitWrite("TLV", tlvs.String())
	t.Log(s.String())
}

func TestTLVMarshalRejectsInvalidLength(t *testing.T) {
	for _, tlv := range []TLV{
		{Tag: MESSAGE_PAYLOAD, Length: 2, ValueBytes: []byte{1}},
		NewTLV(MESSAGE_PAYLOAD, make([]byte, 1<<16)),
	} {
		if _, err := tlv.MarshalBinary(); err == nil {
			t.Fatalf("invalid TLV accepted: length=%d value=%d", tlv.Length, len(tlv.ValueBytes))
		}
	}
}

func TestReadTLVListPreservesRepeatedTags(t *testing.T) {
	first := NewTLV(CALLBACK_NUM, []byte{1, 1, '1'})
	second := NewTLV(CALLBACK_NUM, []byte{1, 1, '2'})
	encoded, err := TLVList{first, second}.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	r := packet.NewPacketReader(encoded)
	defer r.Release()

	got, err := ReadTLVList(r)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || !bytes.Equal(got[0].Value(), first.Value()) || !bytes.Equal(got[1].Value(), second.Value()) {
		t.Fatalf("repeated TLVs not preserved: %#v", got)
	}
	if legacy := got.ToMap()[CALLBACK_NUM]; !bytes.Equal(legacy.Value(), second.Value()) {
		t.Fatalf("legacy map should retain the final occurrence, got %#v", legacy)
	}
}

func TestReadTLVListRejectsTruncation(t *testing.T) {
	tests := [][]byte{
		{0x00},
		{0x00, 0x1e, 0x00},
		{0x00, 0x1e, 0x00, 0x02, '1'},
	}
	for _, data := range tests {
		r := packet.NewPacketReader(data)
		if _, err := ReadTLVList(r); err == nil {
			r.Release()
			t.Fatalf("truncated TLV accepted: %x", data)
		}
		r.Release()
	}
}

func TestTLVMaximumLengthDoesNotOverflow(t *testing.T) {
	value := make([]byte, 1<<16-1)
	tlv := NewTLV(MESSAGE_PAYLOAD, value)
	encoded, err := tlv.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if len(encoded) != 4+len(value) || binary.BigEndian.Uint16(encoded[2:4]) != 0xffff {
		t.Fatalf("unexpected maximum TLV encoding: length=%d header=%d", len(encoded), binary.BigEndian.Uint16(encoded[2:4]))
	}
}

func TestValidateTLVList(t *testing.T) {
	validReceipt := NewTLV(RECEIPTED_MESSAGE_ID, append([]byte("message-id"), 0))
	validSAR := TLVList{
		NewTLV(SAR_MSG_REF_NUM, []byte{0, 1}),
		NewTLV(SAR_TOTAL_SEGMENTS, []byte{2}),
		NewTLV(SAR_SEGMENT_SEQNUM, []byte{1}),
	}
	for _, list := range []TLVList{{validReceipt}, validSAR, {NewTLV(SC_INTERFACE_VERSION, []byte{0x34})}} {
		if err := ValidateTLVList(list); err != nil {
			t.Fatalf("valid TLVs rejected: %v", err)
		}
	}

	invalid := []TLVList{
		{NewTLV(RECEIPTED_MESSAGE_ID, []byte("missing-null"))},
		{NewTLV(SC_INTERFACE_VERSION, []byte{0x35})},
		{NewTLV(SAR_MSG_REF_NUM, []byte{0, 1})},
		{
			NewTLV(SAR_MSG_REF_NUM, []byte{0, 1}),
			NewTLV(SAR_TOTAL_SEGMENTS, []byte{1}),
			NewTLV(SAR_SEGMENT_SEQNUM, []byte{2}),
		},
	}
	for _, list := range invalid {
		if err := ValidateTLVList(list); err == nil {
			t.Fatalf("invalid TLVs accepted: %#v", list)
		}
	}
}
