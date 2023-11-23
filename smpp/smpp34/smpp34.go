package smpp34

import (
	"fmt"
	"time"
	"unicode"
	"unicode/utf8"
)

// IsDeliveryReceipt 根据esmClass判断是否为回执
func IsDeliveryReceipt(esmClass int) bool {
	// x x 0 0 0 0 x x，第5-2位生效，即 0011 1100=0x3c
	// x x 0 0 0 1 x x 表示 回执(Short Message contains SMSC Delivery Receipt)
	return esmClass&0x3c == 0x04
}

// IsLongMO 判断是否为长上行
func IsLongMO(esmClass int) bool {
	// 0 0 x x x x x x，第7-6位生效
	// 0 1 x x x x x x 表示长上行
	return esmClass>>6 == 0x01
}

// GenerateSourceAddress2 解析source_addr 对应的 ton 和 npi
// - 纯数字(包括手机号 或 10DLC)，长码(长度>=10) 用 1/1，短码(长度<10) 用 3/0；
// - 带有字母，用 5/0
// 备注：调研过每家下游供应商对长码的判断都不一样，缺乏参考性。所以我们查询汇总了Byteplus的历史发送，取得 10 这个可信值
func GenerateSourceAddress2(addr string) (ton, npi int, sourceAddress string) {
	if isDigit(addr) {
		if utf8.RuneCountInString(addr) >= 10 {
			return TON_International, NPI_ISDN, addr // 纯数字，长码，1/1
		}
		return TON_NetworkSpecific, NPI_Unknown, addr // 纯数字，短码，3/0
	}
	return TON_Alphanumeric, NPI_Unknown, addr // 带字母，5/0
}

// GenerateSourceAddress 解析source_addr 对应的 ton 和 npi
// 旧协转逻辑.
func GenerateSourceAddress(addr string) (ton, npi int, sourceAddress string) {
	return getTon(addr), 0x00, addr
}

// GenerateDestAddress 解析 dest_address 对应的 ton 和 npi.
// 旧协转逻辑.需要传入得到 addr(目标手机号) 是不带"+"的格式
func GenerateDestAddress(addr string) (ton, npi int, destAddress string) {
	return getTon(addr), 0x01, addr
}

func getTon(address string) (ton int) {
	if isDigit(address) {
		if len(address) >= 10 {
			return TON_International // "E.164 格式" 1
		}
		return TON_Abbreviated // "简短的" 6
	} else {
		return TON_Alphanumeric // "其他格式，数字字母组合" 5
	}
}

// GenerateSourceAddress1 解析 source_address 对应的 ton 和 npi.
// 另一种实现方式
func GenerateSourceAddress1(addr string) (ton, npi int, sourceAddress string) {
	if isLetterOrDigit(addr) {
		l := len(addr)
		if isDigit(addr) /*纯数字*/ {
			// 长度大于 0，一般是"+"开头的手机号
			if l >= 10 {
				return TON_International, NPI_ISDN, addr // ton=1,npi=1
			}
			// 其他 case
			return TON_Abbreviated, NPI_Unknown, addr // ton=6,npi=0
		} else {
			// // short_code
			// if l >= 4 && l <= 6 {
			// 	return TON_NetworkSpecific, NPI_Unknown // ton=3,npi=0
			// }
			return TON_Alphanumeric, NPI_Unknown, addr // ton=5,npi=0
		}
	} else {
		return TON_Unknown, NPI_ISDN, addr // ton=0,npi=1
	}
}

// GenerateDestAddress1 解析 dest_address 对应的 ton 和 npi
// 另一种实现方式
func GenerateDestAddress1(addr string) (ton, npi int, destAddress string) {
	return TON_International, NPI_ISDN, addr
}

// isLetterOrDigit 是否由数字字母组成.
func isLetterOrDigit(s string) bool {
	if len(s) == 0 {
		return false
	}

	for _, i := range s {
		if !unicode.IsLetter(i) && !unicode.IsNumber(i) {
			return false
		}
	}
	return true
}

// isLetter 是否纯字母.
func isLetter(s string) bool {
	if len(s) == 0 {
		return false
	}

	for _, i := range s {
		if !unicode.IsLetter(i) {
			return false
		}
	}
	return true
}

// isDigit 是否纯数字.
func isDigit(s string) bool {
	if len(s) == 0 {
		return false
	}

	for _, i := range s {
		if !unicode.IsDigit(i) {
			return false
		}
	}
	return true
}

// func DecodeSMPP34(data []byte) (PduSMPP, error) {
// 	return ParsePdu(data)
// }

// ToValidatePeriod 将相对时间描述转为 SMPP 规定的时间格式。
func ToValidatePeriod(now time.Time, v string, isRelative bool) (string, error) {
	d, err := time.ParseDuration(v)
	if err != nil {
		return "", fmt.Errorf("parse duration error: %w", err)
	}

	// 不能是 now 之前的
	if now.Add(d).Before(now) {
		return "", fmt.Errorf("invalid target date")
	}

	if isRelative {
		return timeToSMPPTimeFormatRelativce(d), nil
	}
	return timeToSMPPTimeFormatAbsolute(now, now.Add(d)), nil
}

const (
	/*
		请参考：https://bytedance.feishu.cn/file/boxcnqDBHWtrevoJWXQi9JbXhW0  7.1.1.2
			SMPP协议时间格式：YYMMDDhhmmsstnnp，其中：
			‘YY’ 年份后两位(00-99)
			‘MM’ 月份 (01-12)
			‘DD’ 天 (01-31)
			‘hh’ 小时 (00-23)
			‘mm’ 分钟 (00-59)
			‘ss’ 秒 (00-59)
			‘t’ 秒的十分之一 (0-9)
			‘nn’ Time difference in quarter hours between localtime (as expressed in the first 13 octets) and UTC(Universal Time Constant) time (00-48).
			‘p’ - “+” Local time is in quarter hours advanced in relationto UTC time.
			      “-” Local time is in quarter hours retarded in relationto UTC time.
			      “R” Local time is relative to the current SMSC time.    相对时间标志

			当时间是相对时间时：
				t和nn无效，请分别设置为 '0' 和 '00';'p'='R'。
				举例1："020610233429000R" 应该被解释为:
					当前时间的 2 years, 6 months, 10 days, 23 hours, 34minutes and 29 seconds 之后
				举例2："000007000000000R"，应该被解释为：
					当前时间的 7 天之后
	*/
	smppAbsoluteTimeFormat = "060102150405"
	smppRelativeTimeFormat = "0000%02d%02d%02d%02d000R" // 不支持年、月级别的超时时间
)

// timeToSMPPTimeFormatRelativce 将时间t转为SMPP规定的时间格式——相对时间
func timeToSMPPTimeFormatRelativce(diff time.Duration) string {
	days := int(diff.Hours()/24) % 31
	hours := int(diff.Hours()) % 24
	minutes := int(diff.Minutes()) % 60
	seconds := int(diff.Seconds()) % 60
	// 特殊case，全为 0，返回空值
	if days == 0 && hours == 0 && minutes == 0 && seconds == 0 {
		return ""
	}
	return fmt.Sprintf(smppRelativeTimeFormat, days, hours, minutes, seconds)
}

// timeToSMPPTimeFormatAbsolute 将时间t转为SMPP规定的时间格式——绝对时间
func timeToSMPPTimeFormatAbsolute(now, target time.Time) string {
	now, target = now.UTC(), target.UTC() // 最终都用 utc+0 表示
	return target.Format(smppAbsoluteTimeFormat) + "0" + "00" + "+"
}
