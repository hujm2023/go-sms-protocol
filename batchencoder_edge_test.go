package protocol

import (
	"context"
	"testing"

	"github.com/hujm2023/go-sms-protocol/datacoding"
)

func TestBatchDataCodingEncoderRejectsProtocolMismatch(t *testing.T) {
	for _, tt := range []struct {
		name       string
		protocol   Protocol
		dataCoding datacoding.ProtocolDataCoding
	}{
		{
			name:       "CMPP coding on SMPP",
			protocol:   SMPP,
			dataCoding: datacoding.CMPP_CODING_ASCII,
		},
		{
			name:       "SMPP coding on CMPP",
			protocol:   CMPP,
			dataCoding: datacoding.SMPP_CODING_ASCII,
		},
		{
			name:       "unknown protocol",
			protocol:   Protocol("UNKNOWN"),
			dataCoding: datacoding.SMPP_CODING_ASCII,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			contents, actual, err := NewBatchDataCodingEncoder().
				Protocol(tt.protocol).
				Content("hello", 1).
				DataCodings([]datacoding.ProtocolDataCoding{tt.dataCoding}).
				Build(context.Background())
			if err == nil {
				t.Fatal("Build() unexpectedly accepted an invalid protocol/data coding pair")
			}
			if contents != nil {
				t.Fatalf("Build() contents = %v, want nil", contents)
			}
			if actual != datacoding.UnknownProtocolDataCoding {
				t.Fatalf("Build() data coding = %v, want unknown", actual)
			}
		})
	}
}
