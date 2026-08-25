package codec

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"

	"github.com/hujm2023/go-sms-protocol/smpp"
)

type smppCodecReader struct {
	data         []byte
	offset       int
	discardCalls int
	readCalls    int
}

func newSMPPCodecReader(commandLength uint32, payloadLen int) *smppCodecReader {
	data := make([]byte, smpp.MinSMPPHeaderLen+payloadLen)
	binary.BigEndian.PutUint32(data[:smpp.MinSMPPHeaderLen], commandLength)
	for i := smpp.MinSMPPHeaderLen; i < len(data); i++ {
		data[i] = byte(i)
	}
	return &smppCodecReader{data: data}
}

func (r *smppCodecReader) Read(p []byte) (int, error) {
	r.readCalls++
	if r.offset == len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.offset:])
	r.offset += n
	return n, nil
}

func (r *smppCodecReader) Peek(n int) ([]byte, error) {
	if n < 0 {
		return nil, errors.New("negative peek length")
	}
	end := r.offset + n
	if end > len(r.data) {
		return r.data[r.offset:], io.ErrUnexpectedEOF
	}
	return r.data[r.offset:end], nil
}

func (r *smppCodecReader) Discard(n int) (int, error) {
	r.discardCalls++
	if n < 0 || n > r.Size() {
		return 0, io.ErrUnexpectedEOF
	}
	r.offset += n
	return n, nil
}

func (r *smppCodecReader) Size() int {
	return len(r.data) - r.offset
}

func TestSMPPCodecDecodeCommandLength(t *testing.T) {
	codec := NewSMPPCodec()
	tests := []struct {
		name          string
		commandLength uint32
		payloadLen    int
		wantErr       error
		wantDiscard   int
		wantConsumed  int
	}{
		{
			name:          "length zero",
			commandLength: 0,
			wantErr:       smpp.ErrInvalidPudLength,
		},
		{
			name:          "length three",
			commandLength: 3,
			wantErr:       smpp.ErrInvalidPudLength,
		},
		{
			name:          "length four",
			commandLength: smpp.MinSMPPHeaderLen,
			wantErr:       smpp.ErrInvalidPudLength,
		},
		{
			name:          "length fifteen",
			commandLength: 15,
			wantErr:       smpp.ErrInvalidPudLength,
		},
		{
			name:          "incomplete minimum packet",
			commandLength: smpp.MinSMPPPacketLen,
			payloadLen:    smpp.MinSMPPPacketLen - smpp.MinSMPPHeaderLen - 1,
			wantErr:       ErrPacketNotComplete,
		},
		{
			name:          "minimum packet",
			commandLength: smpp.MinSMPPPacketLen,
			payloadLen:    smpp.MinSMPPPacketLen - smpp.MinSMPPHeaderLen,
			wantDiscard:   1,
			wantConsumed:  smpp.MinSMPPPacketLen,
		},
		{
			name:          "maximum packet",
			commandLength: smpp.MAX_PDU_SIZE,
			payloadLen:    smpp.MAX_PDU_SIZE - smpp.MinSMPPHeaderLen,
			wantDiscard:   1,
			wantConsumed:  smpp.MAX_PDU_SIZE,
		},
		{
			name:          "length above maximum",
			commandLength: smpp.MAX_PDU_SIZE + 1,
			wantErr:       smpp.ErrInvalidPudLength,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := newSMPPCodecReader(tt.commandLength, tt.payloadLen)
			got, err := codec.Decode(reader)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Decode() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				want := reader.data[:tt.wantConsumed]
				if !bytes.Equal(got, want) {
					t.Fatalf("Decode() packet = %x, want %x", got, want)
				}
			}
			if reader.discardCalls != tt.wantDiscard {
				t.Fatalf("Discard calls = %d, want %d", reader.discardCalls, tt.wantDiscard)
			}
			if reader.offset != tt.wantConsumed {
				t.Fatalf("consumed bytes = %d, want %d", reader.offset, tt.wantConsumed)
			}
		})
	}
}

func TestSMPPCodecDecodeBlockedCommandLength(t *testing.T) {
	codec := NewSMPPCodec()
	tests := []struct {
		name          string
		commandLength uint32
		payloadLen    int
		wantErr       error
		wantReads     int
		wantConsumed  int
	}{
		{
			name:          "length zero",
			commandLength: 0,
			wantErr:       smpp.ErrInvalidPudLength,
			wantReads:     1,
			wantConsumed:  smpp.MinSMPPHeaderLen,
		},
		{
			name:          "length three",
			commandLength: 3,
			wantErr:       smpp.ErrInvalidPudLength,
			wantReads:     1,
			wantConsumed:  smpp.MinSMPPHeaderLen,
		},
		{
			name:          "length four",
			commandLength: smpp.MinSMPPHeaderLen,
			wantErr:       smpp.ErrInvalidPudLength,
			wantReads:     1,
			wantConsumed:  smpp.MinSMPPHeaderLen,
		},
		{
			name:          "length fifteen",
			commandLength: 15,
			wantErr:       smpp.ErrInvalidPudLength,
			wantReads:     1,
			wantConsumed:  smpp.MinSMPPHeaderLen,
		},
		{
			name:          "incomplete minimum packet",
			commandLength: smpp.MinSMPPPacketLen,
			wantErr:       io.EOF,
			wantReads:     2,
			wantConsumed:  smpp.MinSMPPHeaderLen,
		},
		{
			name:          "minimum packet",
			commandLength: smpp.MinSMPPPacketLen,
			payloadLen:    smpp.MinSMPPPacketLen - smpp.MinSMPPHeaderLen,
			wantReads:     2,
			wantConsumed:  smpp.MinSMPPPacketLen,
		},
		{
			name:          "maximum packet",
			commandLength: smpp.MAX_PDU_SIZE,
			payloadLen:    smpp.MAX_PDU_SIZE - smpp.MinSMPPHeaderLen,
			wantReads:     2,
			wantConsumed:  smpp.MAX_PDU_SIZE,
		},
		{
			name:          "length above maximum",
			commandLength: smpp.MAX_PDU_SIZE + 1,
			wantErr:       smpp.ErrInvalidPudLength,
			wantReads:     1,
			wantConsumed:  smpp.MinSMPPHeaderLen,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := newSMPPCodecReader(tt.commandLength, tt.payloadLen)
			got, err := codec.DecodeBlocked(reader)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("DecodeBlocked() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				want := reader.data[:tt.wantConsumed]
				if !bytes.Equal(got, want) {
					t.Fatalf("DecodeBlocked() packet = %x, want %x", got, want)
				}
			}
			if reader.readCalls != tt.wantReads {
				t.Fatalf("Read calls = %d, want %d", reader.readCalls, tt.wantReads)
			}
			if reader.offset != tt.wantConsumed {
				t.Fatalf("consumed bytes = %d, want %d", reader.offset, tt.wantConsumed)
			}
		})
	}
}
