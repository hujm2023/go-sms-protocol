package exampleutil

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"testing"

	"github.com/cloudwego/netpoll"
	"github.com/hujm2023/go-sms-protocol/smpp"
	"github.com/hujm2023/go-sms-protocol/smpp/smpp34"
)

type byteReader struct{ io.Reader }

func (r byteReader) Read(p []byte) (int, error) { return r.Reader.Read(p[:1]) }

type byteWriter struct{ bytes.Buffer }

func (w *byteWriter) Write(p []byte) (int, error) { return w.Buffer.Write(p[:1]) }

func TestFramingHandlesFragmentedAndCoalescedPDUs(t *testing.T) {
	writer := &byteWriter{}
	request := &smpp34.EnquireLink{Header: smpp.Header{ID: smpp.ENQUIRE_LINK, Sequence: 5}}
	if err := Write(writer, request); err != nil {
		t.Fatal(err)
	}
	frame := append([]byte(nil), writer.Bytes()...)
	reader := byteReader{bytes.NewReader(append(frame, frame...))}
	for i := 0; i < 2; i++ {
		pdu, err := Read(reader, 16, 4096, smpp34.DecodeSMPP34)
		if err != nil || pdu.GetSequenceID() != 5 {
			t.Fatalf("fragmented frame: %v", err)
		}
	}
	buffer := netpoll.NewLinkBuffer()
	defer buffer.Close()
	if _, err := buffer.WriteBinary(append(frame, frame...)); err != nil {
		t.Fatal(err)
	}
	if err := buffer.Flush(); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		pdu, err := Unpack(16, 4096, smpp34.DecodeSMPP34)(context.Background(), buffer)
		if err != nil || pdu.GetSequenceID() != 5 {
			t.Fatalf("coalesced frame: %v", err)
		}
	}
}

func TestFramingRejectsMalformedLengthsBeforeAllocation(t *testing.T) {
	tests := []struct {
		name  string
		frame []byte
	}{
		{name: "zero", frame: []byte{0, 0, 0, 0}},
		{name: "below header", frame: []byte{0, 0, 0, 15}},
		{name: "oversized", frame: []byte{255, 255, 255, 255}},
		{name: "truncated", frame: []byte{0, 0, 0, 16}},
		{name: "incomplete prefix", frame: []byte{0, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Read(bytes.NewReader(tt.frame), 16, 4096, smpp34.DecodeSMPP34); err == nil {
				t.Fatal("malformed frame accepted")
			}
		})
	}
	frame, err := (&smpp34.EnquireLink{Header: smpp.Header{ID: smpp.ENQUIRE_LINK, Sequence: 1}}).IEncode()
	if err != nil {
		t.Fatal(err)
	}
	frame = append(frame, 0)
	binary.BigEndian.PutUint32(frame, uint32(len(frame)))
	if _, err := Read(bytes.NewReader(frame), 16, 4096, smpp34.DecodeSMPP34); err == nil {
		t.Fatal("unexpected PDU body accepted")
	}
}
