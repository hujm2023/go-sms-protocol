package main

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/hujm2023/go-sms-protocol/smgp"
	"github.com/hujm2023/go-sms-protocol/smgp/smgp30"
)

func TestAuthenticateLogin(t *testing.T) {
	now := time.Date(2026, 10, 20, 12, 34, 56, 0, time.UTC)
	cfg := config{account: "900001", password: "test-secret", authSkew: 5 * time.Minute}
	auth := makeClientAuthenticator(cfg.account, cfg.password, 1020123456)
	valid := smgp30.Login{
		ClientID: cfg.account, AuthenticatorClient: string(auth[:]),
		Timestamp: 1020123456, Version: 0x30, LoginMode: smgp.TRANSMIT_MODE,
	}
	tests := []struct {
		name   string
		mutate func(*smgp30.Login)
		want   uint32
	}{
		{name: "valid", want: 0},
		{name: "unknown account", mutate: func(p *smgp30.Login) { p.ClientID = "other" }, want: 21},
		{name: "wrong password", mutate: func(p *smgp30.Login) { p.AuthenticatorClient = string(make([]byte, 16)) }, want: 21},
		{name: "invalid version", mutate: func(p *smgp30.Login) { p.Version = 0x31 }, want: 22},
		{name: "invalid mode", mutate: func(p *smgp30.Login) { p.LoginMode = 3 }, want: 10},
		{name: "stale timestamp", mutate: func(p *smgp30.Login) { p.Timestamp = 1019123456 }, want: 21},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := valid
			if tt.mutate != nil {
				tt.mutate(&request)
			}
			if got := authenticateLogin(&request, cfg, now); got != tt.want {
				t.Fatalf("authentication status = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestLoginAuthenticatorWireVector(t *testing.T) {
	auth := makeClientAuthenticator("900001", "test-secret", 1020123456)
	// Independently calculated from the octets in SMGP §6.2.2 and §6.2.7.
	if got := hex.EncodeToString(auth[:]); got != "1bf6bd9f825047724cb5cb872eea358d" {
		t.Fatalf("client digest = %s", got)
	}
	request := &smgp30.Login{AuthenticatorClient: string(auth[:])}
	request.Header.SequenceID = 42
	response := newLoginResponse(request, 0, "test-secret")
	if got := hex.EncodeToString([]byte(response.AuthenticatorServer)); got != "ffb55ed938726c166993c399dc2ca389" {
		t.Fatalf("server digest = %s", got)
	}
	if response.GetSequenceID() != 42 || response.ServerVersion != 0x30 {
		t.Fatal("login response lost sequence or version")
	}
	if got := newLoginResponse(request, 21, "test-secret"); got.AuthenticatorServer != "" {
		t.Fatal("rejected login contains a server authenticator")
	}
}

func TestLoginTimestampCalendarBoundaries(t *testing.T) {
	tests := []struct {
		name      string
		now       time.Time
		timestamp uint32
		want      bool
	}{
		{name: "year boundary", now: time.Date(2027, 1, 1, 0, 1, 0, 0, time.UTC), timestamp: 1231235900, want: true},
		{name: "leap day", now: time.Date(2028, 2, 29, 0, 1, 0, 0, time.UTC), timestamp: 229000000, want: true},
		{name: "invalid leap day", now: time.Date(2027, 3, 1, 0, 1, 0, 0, time.UTC), timestamp: 229000000},
		{name: "invalid seconds", now: time.Date(2027, 1, 1, 0, 1, 0, 0, time.UTC), timestamp: 101000060},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isFreshLoginTimestamp(tt.timestamp, tt.now, 3*time.Minute); got != tt.want {
				t.Fatalf("timestamp freshness = %t, want %t", got, tt.want)
			}
		})
	}
}
