package packet

import (
	"encoding/hex"
	"testing"
)

const (
	maxPacketReaderFuzzInput = 512
	maxPacketReaderFuzzRead  = 63
)

// FuzzPacketReader exercises bounded operation sequences against Reader. The
// reader's first error is intentionally sticky, so every operation also checks
// that a failed reader cannot consume more input or replace that error.
func FuzzPacketReader(f *testing.F) {
	for _, seed := range []struct {
		data     []byte
		ops      []byte
		readSize uint32
	}{
		{data: []byte("hello\x00world"), ops: []byte{0, 1, 2, 3, 4, 5, 6, 7}, readSize: 5},
		{data: []byte{0xff, 0x00, 0x01}, ops: []byte{0x03, 0x17, 0x2b, 0x3f}, readSize: 3},
		{data: []byte{0x01}, ops: []byte{0x01, 0x09, 0x04, 0x08, 0x00}, readSize: ^uint32(0)},
		{data: nil, ops: []byte{0x00, 0x14, 0x28, 0x3c}, readSize: ^uint32(0)},
		{data: []byte{0x01}, ops: []byte{7}, readSize: ^uint32(0)},
		{data: nil, ops: []byte{7}, readSize: ^uint32(0)},
	} {
		f.Add(seed.data, seed.ops, seed.readSize)
	}

	f.Fuzz(func(t *testing.T, data, ops []byte, readSize uint32) {
		if len(data) > maxPacketReaderFuzzInput {
			data = data[:maxPacketReaderFuzzInput]
		}
		if len(ops) > maxPacketReaderFuzzInput {
			ops = ops[:maxPacketReaderFuzzInput]
		}

		r := NewPacketReader(data)
		for _, rawOp := range ops {
			op := rawOp % 10
			n := int(rawOp>>2) & maxPacketReaderFuzzRead
			beforeErr := r.Error()
			beforeRemaining := r.Remaining()

			switch op {
			case 0:
				r.ReadUint8()
			case 1:
				r.ReadUint16()
			case 2:
				r.ReadUint32()
			case 3:
				r.ReadUint64()
			case 4:
				r.ReadBytes(make([]byte, n))
			case 5:
				r.ReadCStringN(n)
			case 6:
				r.ReadCStringNWithoutTrim(n)
			case 7:
				r.ReadNBytes(int(readSize))
			case 8:
				r.SetErrNil()
			case 9:
				r.Release()
			}

			afterErr := r.Error()
			afterRemaining := r.Remaining()
			if afterRemaining < 0 || afterRemaining > beforeRemaining {
				t.Fatalf("Remaining changed from %d to %d after operation %d", beforeRemaining, afterRemaining, op)
			}
			if beforeErr != nil && op != 8 && op != 9 {
				if afterErr != beforeErr {
					t.Fatalf("operation %d replaced sticky error %v with %v", op, beforeErr, afterErr)
				}
				if afterRemaining != beforeRemaining {
					t.Fatalf("operation %d consumed input after sticky error: %d -> %d", op, beforeRemaining, afterRemaining)
				}
			}

			if beforeErr == nil && afterErr != nil {
				probeRemaining := r.Remaining()
				probeErr := r.Error()
				r.ReadUint8()
				if r.Error() != probeErr || r.Remaining() != probeRemaining {
					t.Fatalf("error was not sticky after operation %d", op)
				}
			}

			if r.Error() == nil {
				remaining := r.Bytes()
				if len(remaining) != r.Remaining() {
					t.Fatalf("Bytes() length = %d, Remaining() = %d", len(remaining), r.Remaining())
				}
				if got, want := r.HexString(), hex.EncodeToString(remaining); got != want {
					t.Fatalf("HexString() = %q, want %q", got, want)
				}
			}
		}
	})
}
