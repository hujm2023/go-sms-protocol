package cmpp

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"strconv"
	"time"
)

const ConnectTSFormat = "0102150405"

// Now ...
func Now() time.Time {
	return time.Now()
}

// GenConnectTimestamp ...
func GenConnectTimestamp(nowFunc func() time.Time) (string, uint32) {
	if nowFunc == nil {
		nowFunc = Now
	}
	t, _ := strconv.Atoi(nowFunc().Format(ConnectTSFormat))
	s := uint32(t)
	return TimeStamp2Str(s), s
}

// TimeStamp2Str converts a timestamp(MMDDHHMMSS) int to a string(10 bytes).
// Right aligned, fill 0 if shorter than 10.
func TimeStamp2Str(t uint32) string {
	return fmt.Sprintf("%010d", t)
}

// GenConnectAuth is used to generate the AuthenticatorSource field in the CMPP CONNECT PDU.
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

// GenConnectRespAuthISMG ...
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
