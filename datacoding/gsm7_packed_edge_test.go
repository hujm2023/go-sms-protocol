package datacoding

import "testing"

func TestGSM7PackedDecodeEmptyIsStable(t *testing.T) {
	for _, source := range []GSM7Packed{nil, GSM7Packed{}} {
		decoded, err := source.Decode()
		if err != nil {
			t.Fatalf("Decode(%v) error = %v", source, err)
		}
		if len(decoded) != 0 {
			t.Fatalf("Decode(%v) = %q, want empty", source, decoded)
		}
	}
}

func TestGSM7PackedPreservesUnambiguousTrailingCR(t *testing.T) {
	const source = "abc\r"

	encoded, err := GSM7Packed(source).Encode()
	if err != nil {
		t.Fatalf("Encode(%q) error = %v", source, err)
	}
	decoded, err := GSM7Packed(encoded).Decode()
	if err != nil {
		t.Fatalf("Decode(%x) error = %v", encoded, err)
	}
	if string(decoded) != source {
		t.Fatalf("Decode(%x) = %q, want %q", encoded, decoded, source)
	}
}
