package smgp30

import (
	"strings"
	"testing"

	"github.com/hujm2023/go-sms-protocol/smgp"
)

func TestLoginAuthenticatorsPreserveBinaryZeros(t *testing.T) {
	for _, position := range []int{0, 7, 15} {
		auth := []byte(strings.Repeat("a", 16))
		auth[position] = 0
		t.Run(string(rune('A'+position)), func(t *testing.T) {
			request := &Login{
				Header:              smgp.NewHeader(0, smgp.CommandLogin, 1),
				AuthenticatorClient: string(auth),
			}
			data, err := request.IEncode()
			if err != nil {
				t.Fatal(err)
			}
			var decoded Login
			if err := decoded.IDecode(data); err != nil {
				t.Fatal(err)
			}
			if decoded.AuthenticatorClient != string(auth) {
				t.Fatalf("client authenticator = %x, want %x", decoded.AuthenticatorClient, auth)
			}
			response := &LoginResp{
				Header:              smgp.NewHeader(0, smgp.CommandLoginResp, 1),
				AuthenticatorServer: string(auth),
			}
			data, err = response.IEncode()
			if err != nil {
				t.Fatal(err)
			}
			var decodedResponse LoginResp
			if err := decodedResponse.IDecode(data); err != nil {
				t.Fatal(err)
			}
			if decodedResponse.AuthenticatorServer != string(auth) {
				t.Fatalf("server authenticator = %x, want %x", decodedResponse.AuthenticatorServer, auth)
			}
		})
	}
}
