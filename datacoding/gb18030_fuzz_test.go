package datacoding

import (
	"bytes"
	"strings"
	"testing"
	"unicode/utf8"
)

const maxGB18030FuzzInput = 512

// FuzzGB18030EncodeDecodeStability covers UTF-8 text, canonical GB18030
// encoding, and arbitrary/malformed GB18030 byte streams.  The x/text
// transformer replacement-decodes malformed input, so arbitrary wire bytes
// are checked for canonical fixed-point behavior instead of byte equality.
func FuzzGB18030EncodeDecodeStability(f *testing.F) {
	for _, seed := range [][]byte{
		{},
		[]byte("Hello world 12345"),
		[]byte("[ByteDance] 我的头发长，天下我为王"),
		[]byte("GB18030: €𠀀"),
		[]byte{0xbb, 0xa8, 0xbc, 0xe4, 0xd2, 0xbb},
		[]byte{0x81},
		[]byte{0xfe, 0x30, 0xff, 0x20},
		[]byte{0x81, 0x30, 0x81, 0x31},
		[]byte{0xff, 0xfe, 0xc3},
		[]byte(strings.Repeat("A", 140)),
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > maxGB18030FuzzInput {
			input = input[:maxGB18030FuzzInput]
		}

		encoded, err := GB18030(input).Encode()
		if err != nil {
			if utf8.Valid(input) {
				t.Fatalf("valid UTF-8 text %q returned GB18030 encode error: %v", input, err)
			}
			if encoded != nil {
				t.Fatalf("failed GB18030 encode returned bytes %x with error %v", encoded, err)
			}
		} else {
			decoded, err := GB18030(encoded).Decode()
			if err != nil {
				t.Fatalf("decoding encoded GB18030 bytes %x: %v", encoded, err)
			}
			if !utf8.Valid(decoded) {
				t.Fatalf("GB18030 decoder returned invalid UTF-8 %x", decoded)
			}
			if utf8.Valid(input) && !bytes.Equal(decoded, input) {
				t.Fatalf("valid UTF-8 text changed: got %q, want %q", decoded, input)
			}
			canonical, err := GB18030(decoded).Encode()
			if err != nil {
				t.Fatalf("re-encoding GB18030 text %q: %v", decoded, err)
			}
			if !bytes.Equal(canonical, encoded) {
				t.Fatalf("GB18030 canonical bytes changed: first %x, second %x", encoded, canonical)
			}
			decodedAgain, err := GB18030(canonical).Decode()
			if err != nil {
				t.Fatalf("decoding canonical GB18030 bytes %x: %v", canonical, err)
			}
			if !bytes.Equal(decodedAgain, decoded) {
				t.Fatalf("GB18030 canonical text changed: first %q, second %q", decoded, decodedAgain)
			}
		}

		rawDecoded, err := GB18030(input).Decode()
		if err != nil {
			t.Fatalf("decoding arbitrary GB18030 bytes %x: %v", input, err)
		}
		if !utf8.Valid(rawDecoded) {
			t.Fatalf("arbitrary GB18030 bytes produced invalid UTF-8 %x", rawDecoded)
		}
		rawCanonical, err := GB18030(rawDecoded).Encode()
		if err != nil {
			t.Fatalf("encoding decoded arbitrary GB18030 bytes %x: %v", input, err)
		}
		rawRoundTrip, err := GB18030(rawCanonical).Decode()
		if err != nil {
			t.Fatalf("decoding re-encoded arbitrary GB18030 bytes %x: %v", rawCanonical, err)
		}
		if !bytes.Equal(rawRoundTrip, rawDecoded) {
			t.Fatalf("arbitrary GB18030 decode/encode/decode changed text: first %q, second %q", rawDecoded, rawRoundTrip)
		}
	})
}
