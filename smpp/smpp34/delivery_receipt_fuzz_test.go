package smpp34

import (
	"strings"
	"testing"
)

const maxDeliveryReceiptFuzzInput = 512

// FuzzDeliveryReceiptRoundTrip parses arbitrary short_message strings and
// checks that the parser/encoder reaches a stable semantic representation.
// IEncode emits a canonical field order, but ExtractDeliveryReceipt does not
// promise to preserve arbitrary whitespace or vendor-specific formatting, so
// the first parse is canonicalized before the fixed-point assertion.
func FuzzDeliveryReceiptRoundTrip(f *testing.F) {
	for _, seed := range []string{
		"",
		"id:1235 sub:001 dlvrd:1 submit date:2107081716 done date:2210131801 stat:DELIVRD err:0 text:aha",
		"id:1235 sub:001 dlvrd:1 done date:2210131801 submit date:2107081716 stat:DELIVRD err:0 text:aha",
		"vendor_id:opaque id:7a44aaba-336f-4a92-9502-dd106aa7369f sub:001 dlvrd:001 submit date:231123193758 done date:231123193800 stat:DELIVRD err:000 text:delivery completed successfully",
		"id:1235 text:hello world stat:DELIVRD",
		"ID:1235 SUB:001 STAT:DELIVRD TEXT:variant",
		"id: stat: text:",
		"id:only",
		"stat:DELIVRD text:",
		"id:1 id:2 stat:DELIVRD err:000 text:duplicate fields",
		"text:before known field stat:DELIVRD",
		"id:1 stat:DELIVRD text:" + strings.Repeat("x", 256),
		string([]byte{'i', 'd', ':', '1', 0, ' ', 's', 't', 'a', 't', ':', 'D', 'E', 'L', 'I', 'V', 'R', 'D', ' ', 't', 'e', 'x', 't', ':', 'x', 0, 0}),
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > maxDeliveryReceiptFuzzInput {
			input = input[:maxDeliveryReceiptFuzzInput]
		}

		extracted, err := ExtractDeliveryReceipt(input)
		if err != nil {
			t.Fatalf("ExtractDeliveryReceipt(%q) returned unexpected error: %v", input, err)
		}
		assertDeliveryReceiptFixedPoint(t, extracted)

		direct := new(DeliveryReceipt)
		if err := direct.IDecode([]byte(input)); err != nil {
			t.Fatalf("IDecode(%q) returned unexpected error: %v", input, err)
		}
		assertDeliveryReceiptFixedPoint(t, *direct)
	})
}

func assertDeliveryReceiptFixedPoint(t *testing.T, receipt DeliveryReceipt) {
	t.Helper()

	encoded, err := receipt.IEncode()
	if err != nil {
		t.Fatalf("IEncode(%+v) returned error: %v", receipt, err)
	}

	canonical := new(DeliveryReceipt)
	if err := canonical.IDecode(encoded); err != nil {
		t.Fatalf("IDecode(IEncode(%+v)) returned error: %v", receipt, err)
	}
	if canonical.Valid() != receipt.Valid() {
		t.Fatalf("Valid changed after semantic round-trip: before %+v, after %+v", receipt, *canonical)
	}

	canonicalBytes, err := canonical.IEncode()
	if err != nil {
		t.Fatalf("IEncode(canonical %+v) returned error: %v", *canonical, err)
	}
	canonicalAgain := new(DeliveryReceipt)
	if err := canonicalAgain.IDecode(canonicalBytes); err != nil {
		t.Fatalf("second IDecode(IEncode(canonical)) returned error: %v", err)
	}
	if *canonicalAgain != *canonical {
		t.Fatalf("receipt semantics did not stabilize: first %+v, second %+v", *canonical, *canonicalAgain)
	}
}
