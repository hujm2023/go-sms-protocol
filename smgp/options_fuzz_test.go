package smgp

import (
	"errors"
	"reflect"
	"testing"

	"github.com/hujm2023/go-sms-protocol/packet"
)

const maxOptionsFuzzInput = 512

// FuzzParseOptions checks both option readers against the same bounded byte
// stream and verifies their shared success and failure contracts.
func FuzzParseOptions(f *testing.F) {
	for _, seed := range [][]byte{
		NewOption(TAG_TP_pid, []byte{0x12, 0x34}).Bytes(),
		append(
			NewOption(TAG_TP_udhi, []byte{1}).Bytes(),
			NewOption(Tag(0xffff), []byte{0xaa, 0xbb}).Bytes()...,
		),
		{},
		nil,
		{0x00, 0x01},
		{0x00, 0x01, 0x00, 0x02, 0xff},
		{0xff, 0xff, 0xff, 0xff},
		{'0', '0', 0x00, 0x00},
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw []byte) {
		if len(raw) > maxOptionsFuzzInput {
			raw = raw[:maxOptionsFuzzInput]
		}

		parsed, parseErr := ParseOptions(raw)
		if parseErr == nil {
			if parsed == nil {
				t.Fatal("ParseOptions returned nil options without an error")
			}

			r := packet.NewPacketReader(raw)
			read := ReadOptions(r)
			if len(raw) == 0 {
				if read != nil || r.Error() != nil || r.Remaining() != 0 {
					t.Fatalf("ReadOptions(empty) = (%v, %v, %d), want (nil, nil, 0)", read, r.Error(), r.Remaining())
				}
			} else {
				if read == nil || r.Error() != nil || r.Remaining() != 0 {
					t.Fatalf("ReadOptions(%x) = (%v, %v, %d), want complete options", raw, read, r.Error(), r.Remaining())
				}
				if !reflect.DeepEqual(read, parsed) {
					t.Fatalf("ReadOptions(%x) = %#v, ParseOptions = %#v", raw, read, parsed)
				}
			}

			serialized := parsed.Serialize()
			if len(serialized) != parsed.Len() {
				t.Fatalf("Serialize length = %d, Options.Len() = %d", len(serialized), parsed.Len())
			}
			reparsed, err := ParseOptions(serialized)
			if err != nil {
				t.Fatalf("ParseOptions(Serialize(%x)) error = %v", raw, err)
			}
			if !reflect.DeepEqual(reparsed, parsed) {
				t.Fatalf("ParseOptions(Serialize(%x)) = %#v, want %#v", raw, reparsed, parsed)
			}
			return
		}

		if !errors.Is(parseErr, ErrLength) {
			t.Fatalf("ParseOptions(%x) error = %v, want ErrLength", raw, parseErr)
		}
		r := packet.NewPacketReader(raw)
		beforeRemaining := r.Remaining()
		if read := ReadOptions(r); read != nil {
			t.Fatalf("ReadOptions(%x) = %#v on malformed input, want nil", raw, read)
		}
		if r.Error() == nil {
			t.Fatalf("ReadOptions(%x) did not preserve a reader error", raw)
		}
		if remaining := r.Remaining(); remaining < 0 || remaining > beforeRemaining {
			t.Fatalf("ReadOptions(%x) remaining = %d, initial = %d", raw, remaining, beforeRemaining)
		}
	})
}
