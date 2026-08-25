package packet

import (
	"errors"
	"testing"
)

func TestPacketReaderNumericTruncationIsSticky(t *testing.T) {
	for _, tt := range []struct {
		name string
		data []byte
		read func(*Reader) uint64
	}{
		{name: "uint8", data: nil, read: func(r *Reader) uint64 { return uint64(r.ReadUint8()) }},
		{name: "uint16", data: []byte{1}, read: func(r *Reader) uint64 { return uint64(r.ReadUint16()) }},
		{name: "uint32", data: []byte{1, 2, 3}, read: func(r *Reader) uint64 { return uint64(r.ReadUint32()) }},
		{name: "uint64", data: []byte{1, 2, 3, 4, 5, 6, 7}, read: func(r *Reader) uint64 { return r.ReadUint64() }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := NewPacketReader(tt.data)
			if got := tt.read(r); got != 0 {
				t.Fatalf("first read = %d, want zero on truncation", got)
			}
			firstErr := r.Error()
			if firstErr == nil {
				t.Fatal("first truncated read did not set an error")
			}
			if got := r.ReadUint8(); got != 0 {
				t.Fatalf("read after error = %d, want zero", got)
			}
			if r.Error() != firstErr {
				t.Fatalf("reader error changed after sticky failure: first=%v, current=%v", firstErr, r.Error())
			}
			if got := r.Remaining(); got != 0 {
				t.Fatalf("Remaining() = %d after consuming truncated input, want zero", got)
			}
		})
	}
}

func TestPacketReaderStringAndBytesTruncationIsSticky(t *testing.T) {
	for _, tt := range []struct {
		name string
		read func(*Reader) string
	}{
		{name: "CStringN", read: func(r *Reader) string { return r.ReadCStringN(3) }},
		{name: "CStringNWithoutTrim", read: func(r *Reader) string { return r.ReadCStringNWithoutTrim(3) }},
		{name: "CString", read: func(r *Reader) string { return r.ReadCString() }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r := NewPacketReader([]byte{'a', 'b'})
			if got := tt.read(r); got != "" {
				t.Fatalf("truncated read = %q, want empty", got)
			}
			firstErr := r.Error()
			if firstErr == nil {
				t.Fatal("truncated string read did not set an error")
			}
			if got := r.ReadCStringN(1); got != "" {
				t.Fatalf("read after error = %q, want empty", got)
			}
			if r.Error() != firstErr {
				t.Fatalf("reader error changed after sticky failure")
			}
		})
	}

	r := NewPacketReader([]byte{1, 2})
	got := r.ReadNBytes(3)
	if got != nil {
		t.Fatalf("ReadNBytes() = %x, want nil on truncation", got)
	}
	if r.Error() == nil {
		t.Fatal("ReadNBytes() truncation did not set an error")
	}

	r = NewPacketReader([]byte{1, 2})
	r.ReadBytes(make([]byte, 3))
	if r.Error() == nil {
		t.Fatal("ReadBytes() truncation did not set an error")
	}
}

func TestPacketReaderRemainingReleaseAndErrorReset(t *testing.T) {
	r := NewPacketReader([]byte{1, 2, 3})
	if got := r.Remaining(); got != 3 {
		t.Fatalf("initial Remaining() = %d, want 3", got)
	}
	if got := r.ReadUint8(); got != 1 {
		t.Fatalf("ReadUint8() = %d, want 1", got)
	}
	if got := r.Remaining(); got != 2 {
		t.Fatalf("Remaining() after read = %d, want 2", got)
	}

	if got := r.ReadNBytes(0); got != nil {
		t.Fatalf("ReadNBytes(0) = %x, want nil", got)
	}
	if got := r.ReadCStringN(0); got != "" {
		t.Fatalf("ReadCStringN(0) = %q, want empty", got)
	}
	if r.Error() != nil {
		t.Fatalf("zero-length reads set an error: %v", r.Error())
	}

	r.ReadUint32()
	if r.Error() == nil {
		t.Fatal("expected a truncation error before Release")
	}
	r.SetErrNil()
	if r.Error() != nil {
		t.Fatalf("SetErrNil() left error %v", r.Error())
	}
	if got := r.Remaining(); got != 0 {
		t.Fatalf("Remaining() after failed read = %d, want 0", got)
	}

	r.Release()
	if r.Error() != nil || r.Remaining() != 0 {
		t.Fatalf("Release() state = error %v, remaining %d; want nil, 0", r.Error(), r.Remaining())
	}
}

func TestPacketReaderErrorRetainsOperationAndCause(t *testing.T) {
	r := NewPacketReader([]byte{1})
	r.ReadUint16()
	err := r.Error()
	if err == nil {
		t.Fatal("expected packet error")
	}
	var opErr *packetOptError
	if !errors.As(err, &opErr) {
		t.Fatalf("Error() type = %T, want *packetOptError", err)
	}
	if opErr.Op() != "ReadNumeric" {
		t.Fatalf("error operation = %q, want ReadNumeric", opErr.Op())
	}
	if opErr.Cause() == nil {
		t.Fatal("packet error has nil cause")
	}
}
