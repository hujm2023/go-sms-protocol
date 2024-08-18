package sgip

import (
	"strings"
	"time"
)

// Timestamp 格式为十进制的mmddhhmmss，比如11月20日20时32分25秒产生的命令，其第二部分为十进制1120203225
func Timestamp(t time.Time) uint32 {
	return uint32(int(t.Month())*100000000 + t.Day()*1000000 +
		t.Hour()*10000 + t.Minute()*100 + t.Second())
}

// FixSGIPMobile sgip协议需要在手机号前面补充86
func FixSGIPMobile(mobile string) string {
	if strings.HasPrefix(mobile, "86") {
		return mobile
	}
	// +86 也返回86
	if strings.HasPrefix(mobile, "+86") {
		return mobile[1:]
	}
	return "86" + mobile
}
