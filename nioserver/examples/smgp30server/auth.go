package main

import (
	"crypto/md5"
	"crypto/subtle"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/hujm2023/go-sms-protocol/smgp/smgp30"
)

// SMGP 3.0.3 §6.2.6 uses these status codes for all responses.
// The historical LoginRespStatus constants describe a different numbering scheme.
const (
	statusOK        uint32 = 0
	statusBusy      uint32 = 1
	statusStructure uint32 = 10
	statusAuth      uint32 = 21
	statusVersion   uint32 = 22
)

func authenticateLogin(request *smgp30.Login, cfg config, now time.Time) uint32 {
	if request.ClientID != cfg.account {
		return statusAuth
	}
	if request.Version != 0x30 {
		return statusVersion
	}
	if request.LoginMode > 2 {
		return statusStructure
	}
	if !isFreshLoginTimestamp(request.Timestamp, now, cfg.authSkew) {
		return statusAuth
	}
	auth := makeClientAuthenticator(cfg.account, cfg.password, request.Timestamp)
	if subtle.ConstantTimeCompare([]byte(request.AuthenticatorClient), auth[:]) != 1 {
		return statusAuth
	}
	return statusOK
}

func makeClientAuthenticator(account, password string, timestamp uint32) [md5.Size]byte {
	data := []byte(account + "\x00\x00\x00\x00\x00\x00\x00" + password + fmt.Sprintf("%010d", timestamp))
	// SMGP protocol exception: 3.0.3 section 6.2.2 requires MD5 for AuthenticatorClient.
	return md5.Sum(data)
}

func newLoginResponse(request *smgp30.Login, status uint32, password string) *smgp30.LoginResp {
	response := request.GenEmptyResponse().(*smgp30.LoginResp)
	response.Status = smgp30.LoginRespStatus(status)
	response.ServerVersion = 0x30
	if status == statusOK {
		// Status is a four-byte network-order integer, not a decimal string.
		data := make([]byte, 4, 4+md5.Size+len(password))
		binary.BigEndian.PutUint32(data, status)
		data = append(data, request.AuthenticatorClient...)
		data = append(data, password...)
		// SMGP protocol exception: 3.0.3 section 6.2.7 requires MD5 for AuthenticatorServer.
		auth := md5.Sum(data)
		response.AuthenticatorServer = string(auth[:])
	}
	return response
}

func isFreshLoginTimestamp(timestamp uint32, now time.Time, allowedSkew time.Duration) bool {
	if allowedSkew <= 0 {
		return false
	}
	// Parse with each actual year so an invalid February 29 is not normalized
	// into March 1; the wire timestamp contains neither a year nor a timezone.
	for _, year := range []int{now.Year() - 1, now.Year(), now.Year() + 1} {
		value := fmt.Sprintf("%04d%010d", year, timestamp)
		candidate, err := time.ParseInLocation("20060102150405", value, now.Location())
		if err != nil {
			continue
		}
		delta := now.Sub(candidate)
		if delta < 0 {
			delta = -delta
		}
		if delta <= allowedSkew {
			return true
		}
	}
	return false
}
