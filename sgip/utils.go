package sgip

import (
	"fmt"
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

func SequenceIDString(sequenceID [3]uint32) string {
	return fmt.Sprintf("%d:%d:%d", sequenceID[0], sequenceID[1], sequenceID[2])
}

func SequenceIDFromString(sequenceID string) [3]uint32 {
	tmp := strings.Split(sequenceID, ":")
	if len(tmp) != 3 {
		return [3]uint32{}
	}
	var id [3]uint32
	_, err := fmt.Sscanf(tmp[0], "%d", &id[0])
	if err != nil {
		return [3]uint32{}
	}
	_, err = fmt.Sscanf(tmp[1], "%d", &id[1])
	if err != nil {
		return [3]uint32{}
	}
	_, err = fmt.Sscanf(tmp[2], "%d", &id[2])
	if err != nil {
		return [3]uint32{}
	}
	return id
}
