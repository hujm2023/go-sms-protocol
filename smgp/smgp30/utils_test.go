package smgp30

import (
	"errors"
	"reflect"
	"testing"

	sms "github.com/hujm2023/go-sms-protocol"
	"github.com/hujm2023/go-sms-protocol/smgp"
)

func TestDecodeSMGP30RoutesSupportedCommands(t *testing.T) {
	tests := []struct {
		name string
		pdu  sms.Encoder
		want interface{}
	}{
		{
			name: "login",
			pdu: &Login{
				Header:              smgp.Header{CommandID: smgp.CommandLogin, SequenceID: 1},
				ClientID:            "client",
				AuthenticatorClient: string(make([]byte, 16)),
			},
			want: (*Login)(nil),
		},
		{name: "login response", pdu: &LoginResp{Header: smgp.Header{CommandID: smgp.CommandLoginResp, SequenceID: 1}}, want: (*LoginResp)(nil)},
		{name: "submit", pdu: &Submit{Header: smgp.Header{CommandID: smgp.CommandSubmit, SequenceID: 1}}, want: (*Submit)(nil)},
		{name: "submit response", pdu: &SubmitResp{Header: smgp.Header{CommandID: smgp.CommandSubmitResp, SequenceID: 1}, MsgID: "01020304050607080901"}, want: (*SubmitResp)(nil)},
		{name: "deliver", pdu: &Deliver{Header: smgp.Header{CommandID: smgp.CommandDeliver, SequenceID: 1}, MsgID: "01020304050607080901"}, want: (*Deliver)(nil)},
		{name: "deliver response", pdu: &DeliverResp{Header: smgp.Header{CommandID: smgp.CommandDeliverResp, SequenceID: 1}, MsgID: "01020304050607080901"}, want: (*DeliverResp)(nil)},
		{name: "active test", pdu: &ActiveTest{Header: smgp.Header{CommandID: smgp.CommandActiveTest, SequenceID: 1}}, want: (*ActiveTest)(nil)},
		{name: "active test response", pdu: &ActiveTestResp{Header: smgp.Header{CommandID: smgp.CommandActiveTestResp, SequenceID: 1}}, want: (*ActiveTestResp)(nil)},
		{name: "exit", pdu: &Exit{Header: smgp.Header{CommandID: smgp.CommandExit, SequenceID: 1}}, want: (*Exit)(nil)},
		{name: "exit response", pdu: &ExitResp{Header: smgp.Header{CommandID: smgp.CommandExitResp, SequenceID: 1}}, want: (*ExitResp)(nil)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.pdu.IEncode()
			if err != nil {
				t.Fatalf("IEncode() error = %v", err)
			}
			got, err := DecodeSMGP30(data)
			if err != nil {
				t.Fatalf("DecodeSMGP30() error = %v", err)
			}
			if reflect.TypeOf(got) != reflect.TypeOf(tt.want) {
				t.Fatalf("DecodeSMGP30() type = %T, want %T", got, tt.want)
			}
		})
	}
}

func TestDecodeSMGP30RejectsUnknownCommand(t *testing.T) {
	data := smgp.NewHeader(smgp.HeaderLength, smgp.CommandID(0x7fffffff), 1).Bytes()

	_, err := DecodeSMGP30(data)
	if !errors.Is(err, sms.ErrUnsupportedPacket) {
		t.Fatalf("DecodeSMGP30() error = %v, want ErrUnsupportedPacket", err)
	}
}
