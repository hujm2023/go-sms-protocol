package consts

import "testing"

func TestProtocolVersionFromStringRejectsMalformedValues(t *testing.T) {
	for _, input := range []string{"", "CMPP", "CMPP_V20_extra", "NOT_A_VERSION"} {
		protocol, version := ProtocolVersionFromString(input)
		if protocol != ProtocolUnknown {
			t.Fatalf("ProtocolVersionFromString(%q) protocol = %q, want unknown", input, protocol)
		}
		if version == nil || version.ToInt() != 0 || version.String() != "unknown" {
			t.Fatalf("ProtocolVersionFromString(%q) version = %#v, want unknown zero version", input, version)
		}
	}
}

func TestProtocolVersionStringHandlesNilAndUnknownVersions(t *testing.T) {
	if got := ProtocolVersionString(ProtocolCMPP, nil); got != "CMPP_0" {
		t.Fatalf("ProtocolVersionString(CMPP, nil) = %q, want CMPP_0", got)
	}
	if got := ProtocolVersionString(ProtocolUnknown, UnknownVersion(0x7f)); got != "UNKNOWN_V7f" {
		t.Fatalf("ProtocolVersionString(unknown, V7f) = %q, want UNKNOWN_V7f", got)
	}
}
