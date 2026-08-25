package codec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"

	"github.com/hujm2023/go-sms-protocol/cmpp"
)

type cmppCodecReader struct {
	data   []byte
	offset int
}

func newCMPPCodecReader(totalLength uint32, payloadLength int) *cmppCodecReader {
	data := make([]byte, cmpp.PacketTotalLengthBytes+payloadLength)
	binary.BigEndian.PutUint32(data[:cmpp.PacketTotalLengthBytes], totalLength)
	for i := cmpp.PacketTotalLengthBytes; i < len(data); i++ {
		data[i] = byte(i)
	}
	return &cmppCodecReader{data: data}
}

func (r *cmppCodecReader) Read(p []byte) (int, error) {
	if r.offset == len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.offset:])
	r.offset += n
	return n, nil
}

func (r *cmppCodecReader) Peek(n int) ([]byte, error) {
	if n < 0 {
		return nil, errors.New("negative peek length")
	}
	end := r.offset + n
	if end > len(r.data) {
		return r.data[r.offset:], io.ErrUnexpectedEOF
	}
	return r.data[r.offset:end], nil
}

func (r *cmppCodecReader) Discard(n int) (int, error) {
	if n < 0 || n > r.Size() {
		return 0, io.ErrUnexpectedEOF
	}
	r.offset += n
	return n, nil
}

func (r *cmppCodecReader) Size() int {
	return len(r.data) - r.offset
}

func TestCMPPCodecDecodeLengthAndStickyPackets(t *testing.T) {
	codec := NewCMPPCodec()
	for _, tt := range []struct {
		name         string
		data         []byte
		wantErr      error
		wantPacket   []byte
		wantConsumed int
	}{
		{
			name:         "partial length field",
			data:         []byte{0, 0, 0},
			wantErr:      ErrPacketNotComplete,
			wantConsumed: 0,
		},
		{
			name:         "declared length zero",
			data:         []byte{0, 0, 0, 0},
			wantErr:      cmpp.ErrInvalidPudLength,
			wantConsumed: 0,
		},
		{
			name:         "declared length one",
			data:         []byte{0, 0, 0, 1},
			wantErr:      cmpp.ErrInvalidPudLength,
			wantConsumed: 0,
		},
		{
			name:         "declared length three",
			data:         []byte{0, 0, 0, 3},
			wantErr:      cmpp.ErrInvalidPudLength,
			wantConsumed: 0,
		},
		{
			name:         "truncated packet",
			data:         newCMPPCodecReader(cmpp.MinCMPPPduLength, cmpp.MinCMPPPduLength-cmpp.PacketTotalLengthBytes-1).data,
			wantErr:      ErrPacketNotComplete,
			wantConsumed: 0,
		},
		{
			name:         "complete packet",
			data:         newCMPPCodecReader(cmpp.MinCMPPPduLength, cmpp.MinCMPPPduLength-cmpp.PacketTotalLengthBytes).data,
			wantPacket:   newCMPPCodecReader(cmpp.MinCMPPPduLength, cmpp.MinCMPPPduLength-cmpp.PacketTotalLengthBytes).data,
			wantConsumed: cmpp.MinCMPPPduLength,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			reader := &cmppCodecReader{data: tt.data}
			got, err := codec.Decode(reader)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Decode() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && !bytes.Equal(got, tt.wantPacket) {
				t.Fatalf("Decode() packet = %x, want %x", got, tt.wantPacket)
			}
			if reader.offset != tt.wantConsumed {
				t.Fatalf("Decode() consumed %d bytes, want %d", reader.offset, tt.wantConsumed)
			}
		})
	}

	first := newCMPPCodecReader(cmpp.MinCMPPPduLength, cmpp.MinCMPPPduLength-cmpp.PacketTotalLengthBytes).data
	second := newCMPPCodecReader(cmpp.MinCMPPPduLength, cmpp.MinCMPPPduLength-cmpp.PacketTotalLengthBytes).data
	reader := &cmppCodecReader{data: append(first, second...)}
	got, err := codec.Decode(reader)
	if err != nil {
		t.Fatalf("Decode() sticky packet error = %v", err)
	}
	if !bytes.Equal(got, first) {
		t.Fatalf("Decode() first packet = %x, want %x", got, first)
	}
	got, err = codec.Decode(reader)
	if err != nil {
		t.Fatalf("Decode() second packet error = %v", err)
	}
	if !bytes.Equal(got, second) {
		t.Fatalf("Decode() second packet = %x, want %x", got, second)
	}
}

func TestCMPPCodecDecodeBlockedLengthAndTruncation(t *testing.T) {
	codec := NewCMPPCodec()
	for _, tt := range []struct {
		name          string
		reader        *cmppCodecReader
		wantErr       error
		wantConsumed  int
		wantPacketLen int
	}{
		{
			name:         "short length field",
			reader:       &cmppCodecReader{data: []byte{0, 0, 0}},
			wantErr:      io.ErrUnexpectedEOF,
			wantConsumed: 3,
		},
		{
			name:         "declared length zero",
			reader:       &cmppCodecReader{data: []byte{0, 0, 0, 0}},
			wantErr:      cmpp.ErrInvalidPudLength,
			wantConsumed: 4,
		},
		{
			name:         "declared length one",
			reader:       &cmppCodecReader{data: []byte{0, 0, 0, 1}},
			wantErr:      cmpp.ErrInvalidPudLength,
			wantConsumed: 4,
		},
		{
			name:         "declared length three",
			reader:       &cmppCodecReader{data: []byte{0, 0, 0, 3}},
			wantErr:      cmpp.ErrInvalidPudLength,
			wantConsumed: 4,
		},
		{
			name: "truncated packet",
			reader: newCMPPCodecReader(cmpp.MinCMPPPduLength,
				cmpp.MinCMPPPduLength-cmpp.PacketTotalLengthBytes-1),
			wantErr:      io.ErrUnexpectedEOF,
			wantConsumed: cmpp.MinCMPPPduLength - 1,
		},
		{
			name: "complete packet",
			reader: newCMPPCodecReader(cmpp.MinCMPPPduLength,
				cmpp.MinCMPPPduLength-cmpp.PacketTotalLengthBytes),
			wantConsumed:  cmpp.MinCMPPPduLength,
			wantPacketLen: cmpp.MinCMPPPduLength,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := codec.DecodeBlocked(tt.reader)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("DecodeBlocked() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantPacketLen > 0 && len(got) != tt.wantPacketLen {
				t.Fatalf("DecodeBlocked() packet length = %d, want %d", len(got), tt.wantPacketLen)
			}
			if tt.reader.offset != tt.wantConsumed {
				t.Fatalf("DecodeBlocked() consumed %d bytes, want %d", tt.reader.offset, tt.wantConsumed)
			}
		})
	}
}
