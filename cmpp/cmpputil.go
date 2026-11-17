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
// Short SourceAddr values are right-padded with zero bytes to 6 bytes before hashing.
// Use [GenConnectAuthUnpadded] if the gateway hashes SourceAddr as supplied.
func GenConnectAuth(account string, password string, timestampStr string) []byte {
	sourceAddr := []byte(account)
	if len(sourceAddr) < 6 {
		sourceAddr = append(sourceAddr, make([]byte, 6-len(sourceAddr))...)
	}

	return genConnectAuth(sourceAddr, password, timestampStr)
}

// GenConnectAuthUnpadded generates AuthenticatorSource without padding SourceAddr.
// Some carriers in mainland China verify AuthenticatorSource this way on
// direct CMPP connections to their own gateways.
// For third-party SMS providers, follow the provider's authentication rules.
//
// The nine zero bytes after SourceAddr are still included in the hash.
// The CONNECT PDU still encodes SourceAddr as a six-byte field.
func GenConnectAuthUnpadded(account string, password string, timestampStr string) []byte {
	return genConnectAuth([]byte(account), password, timestampStr)
}

func genConnectAuth(sourceAddr []byte, password string, timestampStr string) []byte {
	md5Bytes := md5.Sum(
		bytes.Join([][]byte{
			sourceAddr,
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
