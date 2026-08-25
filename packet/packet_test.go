package packet

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type _cmpp2DeliverReqPkt struct {
	MsgId            uint64
	DestId           string // 21
	ServiceId        string // 10
	TpPid            uint8
	TpUdhi           uint8
	MsgFmt           uint8
	SrcTerminalId    string // 21
	RegisterDelivery uint8
	MsgLength        uint8
	MsgContent       string // msgLength
	Reserve          string // 8
}

var (
	b = []byte{181, 37, 98, 128, 0, 1, 0, 0, 57, 48, 48, 48, 48, 49, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 49, 51, 52, 49, 50, 51, 52, 48, 48, 48, 48, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 18, 84, 104, 105, 115, 32, 105, 115, 32, 97, 32, 116, 101, 115, 116, 32, 77, 79, 46, 0, 0, 0, 0, 0, 0, 0, 0}

	dd = &_cmpp2DeliverReqPkt{
		MsgId:            13052947396898652160,
		DestId:           "900001",
		ServiceId:        "",
		TpPid:            0,
		TpUdhi:           0,
		MsgFmt:           0,
		SrcTerminalId:    "13412340000",
		RegisterDelivery: 0,
		MsgLength:        18,
		MsgContent:       "This is a test MO.",
		Reserve:          "",
	}
)

func TestPacketWriter(t *testing.T) {
	t.Run("compare", func(t *testing.T) {
		w := NewPacketWriter(103)
		defer w.Release()

		d := *dd
		w.WriteUint64(d.MsgId)
		w.WriteFixedLenString(d.DestId, 21)
		w.WriteFixedLenString(d.ServiceId, 10)
		w.WriteUint8(d.TpPid)
		w.WriteUint8(d.TpUdhi)
		w.WriteUint8(d.MsgFmt)
		w.WriteFixedLenString(d.SrcTerminalId, 21)
		w.WriteUint8(d.RegisterDelivery)
		w.WriteUint8(d.MsgLength)
		w.WriteString(d.MsgContent)
		w.WriteFixedLenString(d.Reserve, 8)

		data, err := w.Bytes()
		assert.Nil(t, err)

		t.Log(data, w.HexString())
		t.Log(b, hex.EncodeToString(b))
		assert.True(t, bytes.Equal(data, b))
	})

	t.Run("BytesLength", func(t *testing.T) {
		w := NewPacketWriter(0)
		defer w.Release()

		w.WriteUint32(0x03)
		w.WriteUint32(0x00)
		w.WriteUint32(0x01)

		data, err := w.BytesWithLength()
		assert.Nil(t, err)
		assert.Equal(t, int(16), len(data))
		assert.True(t, bytes.Equal(data, []byte{0, 0, 0, 16, 0, 0, 0, 3, 0, 0, 0, 0, 0, 0, 0, 1}))
	})

	t.Run("WriteString", func(t *testing.T) {
		w := NewPacketWriter(0)
		defer w.Release()

		w.WriteCString("")

		t.Log(w.Error())
		t.Log(w.Bytes())

		w.WriteCString("aaaaaa")

		t.Log(w.Error())
		t.Log(w.Bytes())
	})
}

func TestPacketReader(t *testing.T) {
	t.Run("compare", func(t *testing.T) {
		r := NewPacketReader(b)
		defer r.Release()

		d := new(_cmpp2DeliverReqPkt)
		d.MsgId = r.ReadUint64()
		d.DestId = r.ReadCStringN(21)
		d.ServiceId = r.ReadCStringN(10)
		d.TpPid = r.ReadUint8()
		d.TpUdhi = r.ReadUint8()
		d.MsgFmt = r.ReadUint8()
		d.SrcTerminalId = r.ReadCStringN(21)
		d.RegisterDelivery = r.ReadUint8()
		d.MsgLength = r.ReadUint8()

		d.MsgContent = r.ReadCStringN(int(d.MsgLength))
		d.Reserve = r.ReadCStringN(8)

		assert.Nil(t, r.Error())

		assert.True(t, reflect.DeepEqual(d, dd))
	})

	t.Run("ReadCString", func(t *testing.T) {
		bb := []byte{
			0, 0, 0, 38, 0, 0, 0, 9, 0, 0, 0, 0, 0, 0, 0, 1,
			104, 56, 54, 103, 55, 118, 0, 53, 55, 57, 48, 50, 52, 0, 67, 77, 84, 0, 52, 0, 0, 0,
		}

		r := NewPacketReader(bb[16:])
		defer r.Release()

		t.Log(r.ReadCString())
		t.Log(r.ReadCString())
		t.Log(r.ReadCString())
		// t.Log(r.ReadNumeric())
	})
}

func TestPacketWriter_WriteNumeric(t *testing.T) {
	w := NewPacketWriter(0)
	defer w.Release()
	w.WriteUint8(uint8(123))
	assert.Equal(t, 1, w.Len())

	w.WriteFixedLenString("aa", 4)
	assert.Equal(t, 1+4, w.Len())

	w.WriteUint64(uint64(123))
	assert.Equal(t, 1+4+8, w.Len())
}

func TestPacketError(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   string
		limit   int
		wantLen int
	}{
		{name: "ascii", field: "source_addr", value: "secret", limit: 5, wantLen: 6},
		{name: "multibyte octets", field: "服务", value: "短信", limit: 5, wantLen: 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewPacketWriter()
			defer w.Release()

			w.WriteFixedLenStringField(tt.field, tt.value, tt.limit)

			err := w.Error()
			if err == nil {
				t.Fatal("WriteFixedLenStringField accepted an over-limit value")
			}

			var lengthErr *FieldLengthError
			if !errors.As(err, &lengthErr) {
				t.Fatalf("errors.As(%T) did not expose FieldLengthError: %v", err, err)
			}
			assert.Equal(t, tt.field, lengthErr.Field)
			assert.Equal(t, tt.wantLen, lengthErr.Actual)
			assert.Equal(t, tt.limit, lengthErr.Limit)
			assert.ErrorIs(t, err, ErrFieldLengthExceeded)

			fieldText := fmt.Sprintf("field %q", tt.field)
			assert.Contains(t, lengthErr.Error(), fieldText)
			assert.Contains(t, lengthErr.Error(), fmt.Sprintf("actual length %d", tt.wantLen))
			assert.Contains(t, lengthErr.Error(), fmt.Sprintf("limit %d", tt.limit))
			assert.Contains(t, lengthErr.Error(), fmt.Sprintf("excess %d", tt.wantLen-tt.limit))
			assert.NotContains(t, lengthErr.Error(), tt.value)
			assert.NotContains(t, err.Error(), tt.value)

			data, bytesErr := w.Bytes()
			assert.Nil(t, data)
			assert.ErrorIs(t, bytesErr, ErrFieldLengthExceeded)
		})
	}

	t.Run("below and exact limits preserve padding", func(t *testing.T) {
		for _, tt := range []struct {
			name  string
			value string
			want  []byte
		}{
			{name: "below", value: "ab", want: []byte{'a', 'b', 0, 0}},
			{name: "exact", value: "abcd", want: []byte{'a', 'b', 'c', 'd'}},
		} {
			t.Run(tt.name, func(t *testing.T) {
				w := NewPacketWriter()
				defer w.Release()

				w.WriteFixedLenStringField("field", tt.value, 4)
				assert.NoError(t, w.Error())

				data, err := w.Bytes()
				assert.NoError(t, err)
				assert.Equal(t, tt.want, data)
				assert.Equal(t, len(tt.want), w.Written())
				assert.Equal(t, len(tt.want), w.Len())
			})
		}
	})

	t.Run("unnamed legacy method exposes typed overflow", func(t *testing.T) {
		w := NewPacketWriter()
		defer w.Release()

		w.WriteFixedLenString("abcd", 3)

		var lengthErr *FieldLengthError
		if !errors.As(w.Error(), &lengthErr) {
			t.Fatalf("errors.As did not expose FieldLengthError: %v", w.Error())
		}
		assert.Empty(t, lengthErr.Field)
		assert.Equal(t, 4, lengthErr.Actual)
		assert.Equal(t, 3, lengthErr.Limit)
		assert.ErrorIs(t, w.Error(), ErrFieldLengthExceeded)
	})

	t.Run("first error remains sticky and deterministic", func(t *testing.T) {
		w := NewPacketWriter()
		defer w.Release()

		w.WriteFixedLenStringField("first", "abcd", 3)
		firstErr := w.Error()
		if firstErr == nil {
			t.Fatal("expected first write to fail")
		}

		w.WriteFixedLenStringField("second", "efgh", 3)
		w.WriteUint32(1)
		w.WriteBytes([]byte("later"))
		w.WriteString("later")

		assert.Same(t, firstErr, w.Error())
		assert.Equal(t, 0, w.Written())
		assert.Equal(t, 0, w.Len())
		assert.Empty(t, w.HexString())

		data, err := w.Bytes()
		assert.Nil(t, data)
		assert.Same(t, firstErr, err)
		data, err = w.BytesWithLength()
		assert.Nil(t, data)
		assert.Same(t, firstErr, err)
		assert.True(t, strings.Contains(err.Error(), `field "first"`))
	})
}

func BenchmarkReader(bb *testing.B) {
	bb.ResetTimer()
	for i := 0; i < bb.N; i++ {
		r := NewPacketReader(b)

		d := new(_cmpp2DeliverReqPkt)
		d.MsgId = r.ReadUint64()
		d.DestId = r.ReadCStringN(21)
		d.ServiceId = r.ReadCStringN(10)
		d.TpPid = r.ReadUint8()
		d.TpUdhi = r.ReadUint8()
		d.MsgFmt = r.ReadUint8()
		d.SrcTerminalId = r.ReadCStringN(21)
		d.RegisterDelivery = r.ReadUint8()
		d.MsgLength = r.ReadUint8()
		d.MsgContent = string(r.ReadNBytes(int(d.MsgLength)))
		d.Reserve = r.ReadCStringN(8)

		assert.Nil(bb, r.Error())

		assert.True(bb, reflect.DeepEqual(d, dd))

		r.Release()
	}
}

func BenchmarkWriter(bb *testing.B) {
	bb.ResetTimer()
	for i := 0; i < bb.N; i++ {
		r := NewPacketWriter(0)
		r.WriteUint64(dd.MsgId)
		r.WriteFixedLenString(dd.DestId, 21)
		r.WriteFixedLenString(dd.ServiceId, 10)
		r.WriteUint8(dd.TpPid)
		r.WriteUint8(dd.TpUdhi)
		r.WriteUint8(dd.MsgFmt)
		r.WriteFixedLenString(dd.SrcTerminalId, 21)
		r.WriteUint8(dd.RegisterDelivery)
		r.WriteUint8(dd.MsgLength)
		r.WriteFixedLenString(dd.MsgContent, int(dd.MsgLength))

		_, err := r.BytesWithLength()
		assert.Nil(bb, err)

		r.Release()
	}
}
