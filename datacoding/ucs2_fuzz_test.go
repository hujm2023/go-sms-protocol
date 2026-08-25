package datacoding

import (
	"bytes"
	"strings"
	"testing"
	"unicode/utf8"
)

const maxUCS2FuzzInput = 512

// FuzzUCS2EncodeDecodeStability checks the UTF-16BE canonical form used by
// UCS2.  Invalid UTF-8 text and malformed UTF-16 input are replacement-decoded
// by the underlying x/text transformer, so the invariant is canonical
// encode/decode stability rather than equality with malformed source bytes.
func FuzzUCS2EncodeDecodeStability(f *testing.F) {
	for _, seed := range [][]byte{
		{},
		[]byte("Hello world"),
		[]byte("Olá mundão"),
		[]byte("短信内容：你好，世界"),
		[]byte("emoji 🙂"),
		[]byte{0xff, 0xfe, 0xc3},
		[]byte{0x00},
		[]byte{0xd8, 0x3d},
		[]byte{0xdc, 0x00},
		[]byte{0xfe, 0xff, 0x00, 0x41},
		[]byte(strings.Repeat("A", 140)),
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > maxUCS2FuzzInput {
			input = input[:maxUCS2FuzzInput]
		}

		encoded, err := UCS2(input).Encode()
		if err != nil {
			t.Fatalf("encoding UCS2 input %x: %v", input, err)
		}
		decoded, err := UCS2(encoded).Decode()
		if err != nil {
			t.Fatalf("decoding encoded UCS2 bytes %x: %v", encoded, err)
		}
		if !utf8.Valid(decoded) {
			t.Fatalf("UCS2 decoder returned invalid UTF-8 %x", decoded)
		}
		if utf8.Valid(input) && !bytes.Equal(decoded, input) {
			t.Fatalf("valid UTF-8 text changed: got %q, want %q", decoded, input)
		}
		canonical, err := UCS2(decoded).Encode()
		if err != nil {
			t.Fatalf("re-encoding UCS2 text %q: %v", decoded, err)
		}
		if !bytes.Equal(canonical, encoded) {
			t.Fatalf("UCS2 canonical bytes changed: first %x, second %x", encoded, canonical)
		}
		decodedAgain, err := UCS2(canonical).Decode()
		if err != nil {
			t.Fatalf("decoding canonical UCS2 bytes %x: %v", canonical, err)
		}
		if !bytes.Equal(decodedAgain, decoded) {
			t.Fatalf("UCS2 canonical text changed: first %q, second %q", decoded, decodedAgain)
		}

		rawDecoded, err := UCS2(input).Decode()
		if err != nil {
			t.Fatalf("decoding arbitrary UCS2 bytes %x: %v", input, err)
		}
		if !utf8.Valid(rawDecoded) {
			t.Fatalf("arbitrary UCS2 bytes produced invalid UTF-8 %x", rawDecoded)
		}
		rawCanonical, err := UCS2(rawDecoded).Encode()
		if err != nil {
			t.Fatalf("encoding decoded arbitrary UCS2 bytes %x: %v", input, err)
		}
		rawRoundTrip, err := UCS2(rawCanonical).Decode()
		if err != nil {
			t.Fatalf("decoding re-encoded arbitrary UCS2 bytes %x: %v", rawCanonical, err)
		}
		if !bytes.Equal(rawRoundTrip, rawDecoded) {
			t.Fatalf("arbitrary UCS2 decode/encode/decode changed text: first %q, second %q", rawDecoded, rawRoundTrip)
		}
	})
}
