package main

import (
	"crypto/md5"
	"crypto/subtle"
	"fmt"
	"time"

	"github.com/hujm2023/go-sms-protocol/cmpp"
	"github.com/hujm2023/go-sms-protocol/cmpp/cmpp20"
)

// CMPP_CONNECT 的 Timestamp 只有 MMDDHHMMSS，没有年份和时区。
// 服务端按本地时区解析，并在年末/年初选择离当前时间最近的年份。
const connectTimestampLayout = "0102150405"

// authenticateConnect 只负责验证协议字段，不修改连接状态。
// 把状态变更留给调用方，可以保证 CONNECT_RESP 编码成功后才把连接标记为已鉴权。
func authenticateConnect(
	request *cmpp20.PduConnect,
	account string,
	password string,
	now time.Time,
	allowedSkew time.Duration,
) cmpp20.ConnectRespStatus {
	if request.SourceAddr != account {
		return cmpp20.ConnectRespStatusInvalidSourceAddress
	}
	if request.Version != cmpp.Version20 {
		return cmpp20.ConnectRespStatusVersionTooHigh
	}
	if len(request.AuthenticatorSource) != md5.Size {
		return cmpp20.ConnectRespStatusAuthError
	}
	// 时间窗口用于拒绝明显过期的鉴权报文。生产网关还应增加持久化的重放缓存，
	// 因为仅校验时间窗口不能阻止攻击者在窗口内重复发送同一报文。
	if !isFreshConnectTimestamp(request.Timestamp, now, allowedSkew) {
		return cmpp20.ConnectRespStatusAuthError
	}

	want := makeAuthenticatorSource(account, password, request.Timestamp)
	// 鉴权摘要属于秘密相关数据，使用常量时间比较，避免普通字节比较暴露前缀匹配时序。
	if subtle.ConstantTimeCompare([]byte(request.AuthenticatorSource), want[:]) != 1 {
		return cmpp20.ConnectRespStatusAuthError
	}
	return cmpp20.ConnectRespStatusSuccess
}

func makeAuthenticatorSource(account string, password string, timestamp uint32) [md5.Size]byte {
	authenticator := cmpp.GenConnectAuth(account, password, cmpp.TimeStamp2Str(timestamp))
	var result [md5.Size]byte
	copy(result[:], authenticator)
	return result
}

func newConnectResponse(
	request *cmpp20.PduConnect,
	status cmpp20.ConnectRespStatus,
	password string,
) *cmpp20.PduConnectResp {
	response := &cmpp20.PduConnectResp{
		Header:  cmpp.NewHeader(0, cmpp.CommandConnectResp, request.GetSequenceID()),
		Status:  status,
		Version: cmpp.Version20,
	}
	// 协议只要求成功响应携带 AuthenticatorISMG。失败时保持空值，编码器会补齐为 16 个零字节，
	// 避免错误响应泄露任何由密码派生的信息。
	if status == cmpp20.ConnectRespStatusSuccess {
		response.AuthenticatorISMG = string(cmpp.GenConnectRespAuthISMG(
			[]byte{byte(status)},
			request.AuthenticatorSource,
			password,
		))
	}
	return response
}

func isFreshConnectTimestamp(timestamp uint32, now time.Time, allowedSkew time.Duration) bool {
	if allowedSkew <= 0 {
		return false
	}
	parsed, err := time.ParseInLocation(
		connectTimestampLayout,
		fmt.Sprintf("%010d", timestamp),
		now.Location(),
	)
	if err != nil {
		return false
	}

	candidate := nearestYear(parsed, now)
	delta := now.Sub(candidate)
	if delta < 0 {
		delta = -delta
	}
	return delta <= allowedSkew
}

func nearestYear(parsed time.Time, now time.Time) time.Time {
	// Timestamp 不含年份，因此同时比较去年、今年和明年，避免跨年几分钟内误拒绝合法连接。
	best := time.Date(
		now.Year(),
		parsed.Month(),
		parsed.Day(),
		parsed.Hour(),
		parsed.Minute(),
		parsed.Second(),
		0,
		now.Location(),
	)
	bestDelta := now.Sub(best)
	if bestDelta < 0 {
		bestDelta = -bestDelta
	}

	for _, year := range []int{now.Year() - 1, now.Year() + 1} {
		candidate := time.Date(
			year,
			parsed.Month(),
			parsed.Day(),
			parsed.Hour(),
			parsed.Minute(),
			parsed.Second(),
			0,
			now.Location(),
		)
		delta := now.Sub(candidate)
		if delta < 0 {
			delta = -delta
		}
		if delta < bestDelta {
			best = candidate
			bestDelta = delta
		}
	}
	return best
}
