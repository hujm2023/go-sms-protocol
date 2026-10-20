package exampleutil

import (
	"flag"
	"io"
	"testing"
)

func TestServerConfiguration(t *testing.T) {
	tests := []struct {
		name string
		args []string
		ok   bool
	}{
		{name: "defaults", ok: true},
		{name: "zero workers", args: []string{"-receipt-workers=0"}},
		{name: "zero capacity", args: []string{"-receipt-queue=0"}},
		{name: "invalid delay", args: []string{"-receipt-min-delay=1s", "-receipt-max-delay=1ms"}},
		{name: "zero timeout", args: []string{"-write-timeout=0"}},
		{name: "invalid address", args: []string{"-listen=invalid"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg ServerConfig
			f := flag.NewFlagSet("test", flag.ContinueOnError)
			f.SetOutput(io.Discard)
			cfg.Flags(f, "127.0.0.1:0")
			if err := f.Parse(tt.args); err != nil {
				t.Fatal(err)
			}
			if err := cfg.Validate(); (err == nil) != tt.ok {
				t.Fatalf("validation=%v want success=%t", err, tt.ok)
			}
		})
	}
}

func TestCredentialLimitsDoNotExposeValues(t *testing.T) {
	env := func(name string) (string, bool) {
		if name == "TEST_ACCOUNT" {
			return "demo", true
		}
		return "secret", true
	}
	if _, err := LoadCredentials(env, "TEST", 16, 16); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCredentials(env, "TEST", 2, 16); err == nil {
		t.Fatal("long account accepted")
	}
	if _, err := LoadCredentials(func(string) (string, bool) { return "", false }, "TEST", 16, 16); err == nil {
		t.Fatal("missing credentials accepted")
	}
	if err := TextField("a\x00b", 1, 16); err == nil {
		t.Fatal("NUL accepted")
	}
	if err := ASCIIMessage("中文", 140); err == nil {
		t.Fatal("non-ASCII example text accepted")
	}
}
