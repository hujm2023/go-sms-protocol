package main

import (
	"bytes"
	"testing"
	"time"

	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/cmpp/cmpp20"
)

func TestAuthenticateConnect(t *testing.T) {
	const (
		account  = "900001"
		password = "test-secret"
	)
	now := time.Date(2026, time.August, 25, 12, 34, 56, 0, time.UTC)
	validTimestamp := uint32(825123456)
	validAuthenticator := makeAuthenticatorSource(account, password, validTimestamp)

	tests := []struct {
		name           string
		request        cmpp20.PduConnect
		expectedStatus cmpp20.ConnectRespStatus
	}{
		{
			name: "valid credentials",
			request: cmpp20.PduConnect{
				SourceAddr:          account,
				AuthenticatorSource: string(validAuthenticator[:]),
				Version:             cmpp.Version20,
				Timestamp:           validTimestamp,
			},
			expectedStatus: cmpp20.ConnectRespStatusSuccess,
		},
		{
			name: "unknown account",
			request: cmpp20.PduConnect{
				SourceAddr:          "900002",
				AuthenticatorSource: string(validAuthenticator[:]),
				Version:             cmpp.Version20,
				Timestamp:           validTimestamp,
			},
			expectedStatus: cmpp20.ConnectRespStatusInvalidSourceAddress,
		},
		{
			name: "wrong authenticator",
			request: cmpp20.PduConnect{
				SourceAddr:          account,
				AuthenticatorSource: string(make([]byte, 16)),
				Version:             cmpp.Version20,
				Timestamp:           validTimestamp,
			},
			expectedStatus: cmpp20.ConnectRespStatusAuthError,
		},
		{
			name: "unsupported version",
			request: cmpp20.PduConnect{
				SourceAddr:          account,
				AuthenticatorSource: string(validAuthenticator[:]),
				Version:             cmpp.Version30,
				Timestamp:           validTimestamp,
			},
			expectedStatus: cmpp20.ConnectRespStatusVersionTooHigh,
		},
		{
			name: "stale timestamp",
			request: func() cmpp20.PduConnect {
				timestamp := uint32(825120000)
				authenticator := makeAuthenticatorSource(account, password, timestamp)
				return cmpp20.PduConnect{
					SourceAddr:          account,
					AuthenticatorSource: string(authenticator[:]),
					Version:             cmpp.Version20,
					Timestamp:           timestamp,
				}
			}(),
			expectedStatus: cmpp20.ConnectRespStatusAuthError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := authenticateConnect(&test.request, account, password, now, 5*time.Minute)
			if status != test.expectedStatus {
				t.Fatalf("authenticateConnect() = %v, want %v", status, test.expectedStatus)
			}
		})
	}
}

func TestIsFreshConnectTimestampAcrossYearBoundary(t *testing.T) {
	now := time.Date(2027, time.January, 1, 0, 1, 0, 0, time.UTC)
	if !isFreshConnectTimestamp(1231235900, now, 3*time.Minute) {
		t.Fatal("December timestamp immediately before the year boundary was rejected")
	}
}

func TestNewConnectResponse(t *testing.T) {
	requestAuthenticator := []byte{
		0x01, 0x00, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
	}
	request := &cmpp20.PduConnect{
		Header:              cmpp.NewHeader(0, cmpp.CommandConnect, 42),
		AuthenticatorSource: string(requestAuthenticator),
	}

	success := newConnectResponse(request, cmpp20.ConnectRespStatusSuccess, "test-secret")
	wantAuthenticator := cmpp.GenConnectRespAuthISMG(
		[]byte{byte(cmpp20.ConnectRespStatusSuccess)},
		string(requestAuthenticator),
		"test-secret",
	)
	if !bytes.Equal([]byte(success.AuthenticatorISMG), wantAuthenticator) {
		t.Fatalf("AuthenticatorISMG = %x, want %x", success.AuthenticatorISMG, wantAuthenticator)
	}
	if success.GetSequenceID() != request.GetSequenceID() || success.Version != cmpp.Version20 {
		t.Fatalf("response header/version = sequence %d version %#x", success.GetSequenceID(), success.Version)
	}

	rejected := newConnectResponse(request, cmpp20.ConnectRespStatusAuthError, "test-secret")
	if rejected.AuthenticatorISMG != "" {
		t.Fatalf("rejected AuthenticatorISMG = %x, want empty", rejected.AuthenticatorISMG)
	}
}
