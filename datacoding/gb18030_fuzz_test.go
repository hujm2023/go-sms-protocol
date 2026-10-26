package datacoding

import (
	"bytes"
	"strings"
	"testing"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
)

const maxGB18030FuzzInput = 512

// FuzzGB18030EncodeDecodeStability covers UTF-8 text, canonical GB18030
// encoding, and arbitrary/malformed GB18030 byte streams.  The x/text
// transformer replacement-decodes malformed input, so arbitrary wire bytes
// are checked for canonical fixed-point behavior instead of byte equality.
// Some valid private-use runes also do not round-trip through x/text
// (https://go.dev/issue/62063). Check text conversion against that dependency
// instead of assuming every valid UTF-8 input is preserved.
func FuzzGB18030EncodeDecodeStability(f *testing.F) {
	for _, seed := range [][]byte{
		{},
		[]byte("Hello world 12345"),
		[]byte("[ByteDance] 我的头发长，天下我为王"),
		[]byte("GB18030: €𠀀"),
		[]byte("\ue805"),
		[]byte("\ue000\ue7c7\ue864"),
		[]byte("中文\ue805€𠀀"),
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
			referenceEncoded, err := simplifiedchinese.GB18030.NewEncoder().Bytes(input)
			if err != nil {
				t.Fatalf("reference GB18030 encode failed: %v", err)
			}
			if !bytes.Equal(encoded, referenceEncoded) {
				t.Fatalf("GB18030 encoding differs from x/text: got %x, want %x", encoded, referenceEncoded)
			}
			referenceDecoded, err := simplifiedchinese.GB18030.NewDecoder().Bytes(referenceEncoded)
			if err != nil {
				t.Fatalf("reference GB18030 decode failed: %v", err)
			}
			if !bytes.Equal(decoded, referenceDecoded) {
				t.Fatalf("GB18030 decoded text differs from x/text: got %q, want %q", decoded, referenceDecoded)
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
