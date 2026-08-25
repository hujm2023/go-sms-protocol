package datacoding

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	gsm7 "github.com/hujm2023/go-sms-protocol/datacoding/gsm7encoding"
)

const maxGSM7UnpackedFuzzInput = 512

// FuzzGSM7UnpackedRoundTrip exercises both the UTF-8 text and unpacked GSM7
// wire representations.  Unpacked GSM7 has a one-to-one canonical mapping
// for accepted septets, while malformed text or septets have explicit errors
// and are returned unchanged by the codec wrapper.
func FuzzGSM7UnpackedRoundTrip(f *testing.F) {
	for _, seed := range [][]byte{
		{},
		[]byte("Hello world"),
		[]byte("^{}\\[~]|€"),
		[]byte("@£$¥èéùìòÇ\nØø\rÅåΔ_ΦΓΛΩΠΨΣΘΞÆæßÉ"),
		[]byte("你"),
		[]byte{0xff, 0xfe},
		[]byte{0x1b},
		[]byte{0x1b, 0x00},
		[]byte{0x80},
		[]byte(strings.Repeat("A", 160)),
		bytes.Repeat([]byte{0x1b, 0x3c}, 80),
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > maxGSM7UnpackedFuzzInput {
			input = input[:maxGSM7UnpackedFuzzInput]
		}

		text := string(input)
		validText := CanEncodeByGSM7(text)
		encoded, err := GSM7Unpacked(input).Encode()
		if err != nil {
			if validText {
				t.Fatalf("valid GSM7 text %q returned encode error: %v", text, err)
			}
			if !errors.Is(err, gsm7.ErrInvalidCharacter) {
				t.Fatalf("invalid GSM7 text %x returned unexpected error: %v", input, err)
			}
			if !bytes.Equal(encoded, input) {
				t.Fatalf("failed encode changed input: got %x, want %x", encoded, input)
			}
		} else {
			if !validText {
				t.Fatalf("invalid GSM7 text %x unexpectedly encoded as %x", input, encoded)
			}
			decoded, err := GSM7Unpacked(encoded).Decode()
			if err != nil {
				t.Fatalf("decoding encoded GSM7 text %q: %v", text, err)
			}
			if !bytes.Equal(decoded, input) {
				t.Fatalf("text round-trip changed input: got %q, want %q", decoded, input)
			}
			reencoded, err := GSM7Unpacked(decoded).Encode()
			if err != nil {
				t.Fatalf("re-encoding decoded GSM7 text %q: %v", decoded, err)
			}
			if !bytes.Equal(reencoded, encoded) {
				t.Fatalf("canonical GSM7 bytes changed: first %x, second %x", encoded, reencoded)
			}
		}

		validWire := len(gsm7.ValidateGSM7Buffer(input)) == 0
		decoded, err := GSM7Unpacked(input).Decode()
		if err != nil {
			if validWire {
				t.Fatalf("valid GSM7 bytes %x returned decode error: %v", input, err)
			}
			if !errors.Is(err, gsm7.ErrInvalidByte) {
				t.Fatalf("invalid GSM7 bytes %x returned unexpected error: %v", input, err)
			}
			if !bytes.Equal(decoded, input) {
				t.Fatalf("failed decode changed input: got %x, want %x", decoded, input)
			}
			return
		}
		if !validWire {
			t.Fatalf("invalid GSM7 bytes %x unexpectedly decoded as %q", input, decoded)
		}
		reencoded, err := GSM7Unpacked(decoded).Encode()
		if err != nil {
			t.Fatalf("encoding decoded GSM7 bytes %x: %v", input, err)
		}
		if !bytes.Equal(reencoded, input) {
			t.Fatalf("GSM7 wire round-trip changed bytes: got %x, want %x", reencoded, input)
		}
	})
}
