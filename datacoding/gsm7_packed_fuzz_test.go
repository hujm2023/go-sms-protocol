package datacoding

import (
	"bytes"
	"testing"
)

const maxGSM7PackedFuzzInput = 256

// FuzzGSM7PackedDecodeEncodeStability verifies that every packed byte stream
// accepted by the decoder can be canonicalized to a stable wire form. Packed
// bytes do not carry a septet count, so an all-zero final septet is inherently
// ambiguous and may be omitted by Unpack; the fixed-point check allows that
// documented canonicalization while still catching unstable encode/decode
// behavior. Inputs are bounded because packed data expands into septets and
// UTF-8 output.
func FuzzGSM7PackedDecodeEncodeStability(f *testing.F) {
	for _, seed := range [][]byte{
		{0xc8, 0x32, 0x9b, 0xfd, 0x06, 0xdd, 0xdf, 0x72, 0x36, 0x19}, // Hello world
		{0x1b, 0xca, 0x06, 0xb5, 0x49, 0x6d, 0x5e, 0x1b, 0xde, 0xa6, 0xb7, 0xf1, 0x6d, 0x80, 0x9b, 0x32},
		{0x31, 0xd9, 0x8c, 0x56, 0xb3, 0xdd, 0x1a}, // seven septets and filler handling
		{},
		{0x1b},       // truncated extension escape
		{0x1b, 0x80}, // invalid extension byte
		{0xff, 0xff},
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, packed []byte) {
		if len(packed) > maxGSM7PackedFuzzInput {
			packed = packed[:maxGSM7PackedFuzzInput]
		}

		decoded, err := GSM7Packed(packed).Decode()
		if err != nil {
			if decoded != nil {
				t.Fatalf("invalid packed bytes %x returned decoded data %q", packed, decoded)
			}
			return
		}

		canonical, err := GSM7Packed(decoded).Encode()
		if err != nil {
			t.Fatalf("re-encoding decoded text %q: %v", decoded, err)
		}
		initialCanonical := append([]byte(nil), canonical...)
		// A packed stream cannot distinguish a trailing zero septet from
		// filler, so canonicalization may need one extra pass to settle on
		// the filler form. Each non-stable pass removes at least one
		// ambiguous trailing octet; require convergence within the input's
		// bounded canonical length.
		for pass := 0; pass <= len(initialCanonical); pass++ {
			roundTrip, err := GSM7Packed(canonical).Decode()
			if err != nil {
				t.Fatalf("decoding canonical bytes %x: %v", canonical, err)
			}
			canonicalAgain, err := GSM7Packed(roundTrip).Encode()
			if err != nil {
				t.Fatalf("re-encoding round-trip text %q: %v", roundTrip, err)
			}
			if bytes.Equal(canonicalAgain, canonical) {
				return
			}
			canonical = canonicalAgain
		}
		t.Fatalf("canonical encoding did not converge from %x", initialCanonical)
	})
}
