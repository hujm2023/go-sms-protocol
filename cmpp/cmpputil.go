package cmpp

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"strconv"
	"time"
)

// ConnectTSFormat defines the timestamp format (MMDDHHMMSS) used in CMPP Connect PDU.
const ConnectTSFormat = "0102150405"

// Now returns the current time. It's a variable for easy mocking in tests.
var Now = time.Now

// GenConnectTimestamp generates the timestamp string and uint32 value for CMPP Connect PDU.
// It uses the format defined by ConnectTSFormat.
func GenConnectTimestamp(nowFunc func() time.Time) (string, uint32) {
	if nowFunc == nil {
		nowFunc = Now
	}
	t, _ := strconv.Atoi(nowFunc().Format(ConnectTSFormat))
	s := uint32(t)
	return TimeStamp2Str(s), s
}

// TimeStamp2Str converts a timestamp (MMDDHHMMSS format uint32) to a 10-byte string.
// It pads with leading zeros if necessary.
func TimeStamp2Str(t uint32) string {
	return fmt.Sprintf("%010d", t)
}

// GenConnectAuth generates the AuthenticatorSource field for the CMPP CONNECT PDU.
// It calculates the MD5 hash of (SourceAddr + 9 bytes of zeros + password + Timestamp).
func GenConnectAuth(account string, password string, timestampStr string) []byte {
	md5Bytes := md5.Sum(
		bytes.Join([][]byte{
			[]byte(account),
			make([]byte, 9),
			[]byte(password),
			[]byte(timestampStr),
		},
			nil),
	)
	return md5Bytes[:]
}

// GenConnectRespAuthISMG generates the AuthenticatorISMG field for the CMPP CONNECT_RESP PDU.
// It calculates the MD5 hash of (Status + AuthenticatorSource + password).
func GenConnectRespAuthISMG(statusBytes []byte, reqAuth string, password string) []byte {
	m := md5.Sum(bytes.Join([][]byte{
		statusBytes,
		[]byte(reqAuth),
		[]byte(password),
	},
		nil),
	)
	return m[:]
}
