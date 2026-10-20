package main

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/hujm2023/hlog"
)

func testEnvironment(name string) (string, bool) {
	switch name {
	case accountEnvironment:
		return "900001", true
	case passwordEnvironment:
		return "test-secret", true
	default:
		return "", false
	}
}

func TestParseConfigRejectsUnsafeSettings(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "gateway overflow", args: []string{"-gateway-id=1000000"}},
		{name: "queue smaller than one submit", args: []string{"-receipt-queue=99"}},
		{name: "zero workers", args: []string{"-receipt-workers=0"}},
		{name: "negative skew", args: []string{"-auth-skew=-1s"}},
		{name: "inverted receipt delays", args: []string{"-receipt-min-delay=1s", "-receipt-max-delay=1ms"}},
		{name: "zero heartbeats", args: []string{"-heartbeat-interval=0"}},
		{name: "zero heartbeat attempts", args: []string{"-heartbeat-misses=0"}},
		{name: "zero read timeout", args: []string{"-read-timeout=0"}},
		{name: "zero shutdown timeout", args: []string{"-shutdown-timeout=0"}},
		{name: "empty listen", args: []string{"-listen="}},
		{name: "unexpected args", args: []string{"extra"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := parseConfig(tt.args, testEnvironment, io.Discard); err == nil {
				t.Fatal("invalid configuration accepted")
			}
		})
	}
	if _, err := parseConfig(nil, testEnvironment, io.Discard); err != nil {
		t.Fatalf("defaults rejected: %v", err)
	}
	if _, err := parseConfig(nil, func(string) (string, bool) { return "", false }, io.Discard); err == nil {
		t.Fatal("missing credentials accepted")
	}
	if code := realMain([]string{"-help"}, testEnvironment, io.Discard); code != 0 {
		t.Fatalf("help exit code = %d", code)
	}
	if code := realMain([]string{"-unknown"}, testEnvironment, io.Discard); code != 2 {
		t.Fatalf("invalid flags exit code = %d", code)
	}
}

func TestRunRejectsInvalidListenerWithoutLeakingWorkers(t *testing.T) {
	cfg, err := parseConfig(nil, testEnvironment, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	cfg.listenAddress = "invalid-address"
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := run(ctx, cfg, hlog.DefaultLogger()); err == nil {
		t.Fatal("invalid listener did not report startup failure")
	}
}
