package main

import (
	"io"
	"strings"
	"testing"
)

func TestConfigurationExtendedCredentials(t *testing.T) {
	account := strings.Repeat("a", 128)
	password := strings.Repeat("p", 256)
	lookup := func(name string) (string, bool) {
		switch name {
		case "SMPP34_ACCOUNT":
			return account, true
		case "SMPP34_PASSWORD":
			return password, true
		default:
			return "", false
		}
	}
	cfg, err := parseConfig(nil, lookup, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Account != account || cfg.Password != password {
		t.Fatal("credentials changed during configuration")
	}
}
