package main

import (
	"bytes"
	"encoding/hex"
	"io"
	"strings"
	"testing"

	"github.com/hujm2023/go-sms-protocol/smgp"
	"github.com/hujm2023/go-sms-protocol/smgp/smgp30"
)

func TestReadPDUHandlesFragmentationAndRejectsInvalidSizes(t *testing.T) {
	data, err := (&smgp30.ActiveTest{Header: smgp.NewHeader(0, smgp.CommandActiveTest, 5)}).IEncode()
	if err != nil {
		t.Fatal(err)
	}
	// A byte-at-a-time reader exercises fragmented headers and bodies.
	reader := &oneByteReader{Reader: bytes.NewReader(append(data, data...))}
	for i := 0; i < 2; i++ {
		pdu, err := readPDU(reader)
		if err != nil || pdu.GetSequenceID() != 5 {
			t.Fatalf("fragmented/coalesced frame: %v", err)
		}
	}
	for _, data := range [][]byte{{0, 0, 0, 0}, {0xff, 0xff, 0xff, 0xff}, {0, 0, 0, 12}} {
		if _, err := readPDU(bytes.NewReader(data)); err == nil {
			t.Fatal("invalid/truncated frame accepted")
		}
	}
}

type oneByteReader struct{ io.Reader }

func (r *oneByteReader) Read(p []byte) (int, error) { return r.Reader.Read(p[:1]) }

type oneByteWriter struct{ bytes.Buffer }

func (w *oneByteWriter) Write(p []byte) (int, error) { return w.Buffer.Write(p[:1]) }

func TestWritePDUHandlesShortWrites(t *testing.T) {
	writer := &oneByteWriter{}
	pdu := &smgp30.ActiveTest{Header: smgp.NewHeader(0, smgp.CommandActiveTest, 6)}
	if err := writePDU(writer, pdu); err != nil {
		t.Fatal(err)
	}
	decoded, err := readPDU(&writer.Buffer)
	if err != nil || decoded.GetSequenceID() != 6 {
		t.Fatalf("short-write round trip failed: %v", err)
	}
}

func TestVerifyLoginResponseUsesFourByteStatus(t *testing.T) {
	clientAuth, err := hex.DecodeString("1bf6bd9f825047724cb5cb872eea358d")
	if err != nil {
		t.Fatal(err)
	}
	serverAuth, err := hex.DecodeString("ffb55ed938726c166993c399dc2ca389")
	if err != nil {
		t.Fatal(err)
	}
	login := &smgp30.Login{AuthenticatorClient: string(clientAuth)}
	response := &smgp30.LoginResp{AuthenticatorServer: string(serverAuth), ServerVersion: 0x30}
	if err := verifyLoginResponse(login, response, "test-secret"); err != nil {
		t.Fatal(err)
	}
	if err := verifyLoginResponse(login, response, "wrong-secret"); err == nil {
		t.Fatal("invalid server authenticator accepted")
	}
	response.Status = 21
	if err := verifyLoginResponse(login, response, "test-secret"); err == nil {
		t.Fatal("rejected login treated as authenticated")
	}
}

func TestClientConfig(t *testing.T) {
	env := func(name string) (string, bool) {
		if name == "SMGP30_ACCOUNT" {
			return "900001", true
		}
		return "test-secret", true
	}
	if _, err := parseConfig(nil, env, io.Discard); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"-timeout=0"}, {"-message=中文"}, {"-message=" + strings.Repeat("a", 141)},
		{"-source="}, {"-destination="}, {"extra"},
	} {
		if _, err := parseConfig(args, env, io.Discard); err == nil {
			t.Fatalf("invalid client arguments accepted: %q", args)
		}
	}
	if code := realMain([]string{"-help"}, env, io.Discard, io.Discard); code != 0 {
		t.Fatalf("help exit code = %d", code)
	}
	if code := realMain([]string{"-timeout=0"}, env, io.Discard, io.Discard); code != 2 {
		t.Fatalf("configuration error exit code = %d", code)
	}
}
