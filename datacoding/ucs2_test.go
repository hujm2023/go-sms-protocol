package datacoding

import (
	"bytes"
	"testing"
)

func TestUCS2(t *testing.T) {
	want := []byte("\x00O\x00l\x00\xe1\x00 \x00m\x00u\x00n\x00d\x00\xe3\x00o")
	text := []byte("Olá mundão")
	s := UCS2(text)
	have, _ := s.Encode()
	if !bytes.Equal(want, have) {
		t.Fatalf("Unexpected text; want %q, have %q", want, have)
	}
}
