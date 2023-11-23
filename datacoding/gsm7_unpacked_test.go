package datacoding

import (
	"bytes"
	"testing"
)

func TestGSM7UnPacked(t *testing.T) {
	want := []byte("\x48\x65\x6C\x6C\x6F \x77\x6F\x72\x6C\x64")
	text := []byte("Hello world")
	s := GSM7Unpacked(text)
	have, _ := s.Encode()
	if !bytes.Equal(want, have) {
		t.Fatalf("Unexpected text; want %q, have %q", want, have)
	}
}
