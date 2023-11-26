package smpp34

import (
	"fmt"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/hujm2023/go-sms-protocol/smpp"
)

// IsDeliveryReceipt determines if it is a delivery receipt
func IsDeliveryReceipt(esmClass int) bool {
	// x x 0 0 0 0 x x，第5-2位生效，即 0011 1100=0x3c
	// x x 0 0 0 1 x x 表示 回执(Short Message contains SMSC Delivery Receipt)
	return esmClass&0x3c == 0x04
}

// IsLongMO determines if it is a long (multipart) incoming message.
func IsLongMO(esmClass int) bool {
	// 0 0 x x x x x x，第7-6位生效
	// 0 1 x x x x x x 表示长上行
	return esmClass>>6 == 0x01
}

// GenerateSourceAddress2 parses the ton (Type of Number) and npi (Numbering Plan Indicator)
// corresponding to the source_addr:
//   - Pure numbers (including mobile numbers or 10DLC) and long codes (length >= 10) use 1/1,
//     short codes (length < 10) use 3/0.
//   - If the source_addr contains letters, it uses 5/0.
//
// Note: Research shows that each downstream provider has different criteria for determining
// long codes, which lacks reference. Here, we make a decision and set it to 10.
func GenerateSourceAddress2(addr string) (ton, npi int, sourceAddress string) {
	if isDigit(addr) {
		if utf8.RuneCountInString(addr) >= 10 {
			return smpp.TON_International, smpp.NPI_ISDN, addr // 纯数字，长码，1/1
		}
		return smpp.TON_NetworkSpecific, smpp.NPI_Unknown, addr // 纯数字，短码，3/0
	}
	return smpp.TON_Alphanumeric, smpp.NPI_Unknown, addr // 带字母，5/0
}

// GenerateSourceAddress ...
// Deprecated(use GenerateSourceAddress2).
func GenerateSourceAddress(addr string) (ton, npi int, sourceAddress string) {
	return getTon(addr), 0x00, addr
}

// GenerateDestAddress ...
func GenerateDestAddress(addr string) (ton, npi int, destAddress string) {
	return getTon(addr), 0x01, addr
}

func getTon(address string) (ton int) {
	if isDigit(address) {
		if len(address) >= 10 {
			return smpp.TON_International // "E.164 格式" 1
		}
		return smpp.TON_Abbreviated // "简短的" 6
	} else {
		return smpp.TON_Alphanumeric // "其他格式，数字字母组合" 5
	}
}

// GenerateSourceAddress1 ...
// Deprecated(use GenerateSourceAddress2).
func GenerateSourceAddress1(addr string) (ton, npi int, sourceAddress string) {
	if isLetterOrDigit(addr) {
		l := len(addr)
		if isDigit(addr) /*纯数字*/ {
			// 长度大于 0，一般是"+"开头的手机号
			if l >= 10 {
				return smpp.TON_International, smpp.NPI_ISDN, addr // ton=1,npi=1
			}
			// 其他 case
			return smpp.TON_Abbreviated, smpp.NPI_Unknown, addr // ton=6,npi=0
		} else {
			// // short_code
			// if l >= 4 && l <= 6 {
			// 	return TON_NetworkSpecific, NPI_Unknown // ton=3,npi=0
			// }
			return smpp.TON_Alphanumeric, smpp.NPI_Unknown, addr // ton=5,npi=0
		}
	} else {
		return smpp.TON_Unknown, smpp.NPI_ISDN, addr // ton=0,npi=1
	}
}

// GenerateDestAddress1 ...
// 另一种实现方式
func GenerateDestAddress1(addr string) (ton, npi int, destAddress string) {
	return smpp.TON_International, smpp.NPI_ISDN, addr
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

// ToValidatePeriod converts a relative time description to the time format specified by the SMPP protocol.
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
