package main

import (
	"testing"

	"github.com/hujm2023/go-sms-protocol/cmpp/cmpp20"
)

func TestValidateSubmit(t *testing.T) {
	valid := cmpp20.PduSubmit{
		PkTotal:            1,
		PkNumber:           1,
		RegisteredDelivery: 1,
		DestUsrTL:          1,
		DestTerminalID:     []string{"13800138000"},
		MsgLength:          2,
		MsgContent:         []byte("ok"),
	}
	tests := []struct {
		name           string
		mutate         func(*cmpp20.PduSubmit)
		expectedResult uint8
	}{
		{name: "valid submit", expectedResult: 0},
		{
			name: "invalid packet number",
			mutate: func(request *cmpp20.PduSubmit) {
				request.PkNumber = 2
			},
			expectedResult: 1,
		},
		{
			name: "recipient count mismatch",
			mutate: func(request *cmpp20.PduSubmit) {
				request.DestUsrTL = 2
			},
			expectedResult: 1,
		},
		{
			name: "invalid registered delivery",
			mutate: func(request *cmpp20.PduSubmit) {
				request.RegisteredDelivery = 3
			},
			expectedResult: 1,
		},
		{
			name: "message length mismatch",
			mutate: func(request *cmpp20.PduSubmit) {
				request.MsgLength = 3
			},
			expectedResult: 4,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := valid
			request.DestTerminalID = append([]string(nil), valid.DestTerminalID...)
			request.MsgContent = append([]byte(nil), valid.MsgContent...)
			if test.mutate != nil {
				test.mutate(&request)
			}
			if result := validateSubmit(&request); result != test.expectedResult {
				t.Fatalf("validateSubmit() = %d, want %d", result, test.expectedResult)
			}
		})
	}
}
