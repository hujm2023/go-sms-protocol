package datacoding

import (
	"math"

	"github.com/pkg/errors"
)

const (
	UDHILength       = 6
	MaxLongSmsLength = 140                            // 非 gsm7_unpacked 编码 长短信最大字符
	MaxGSM7Length    = 160                            // gsm7_unpacked 编码 长短信最大长度
	SplitBy134       = MaxLongSmsLength - UDHILength  // 非 gsm7_unpacked 编码 长短信切割标准
	SplitBy153       = MaxGSM7Length - UDHILength - 1 // gsm7_unpacked 编码 长短信切割标准
	// 为什么gsm4_unpacked的切割标准是 160-7=153，而不是160-6=154？
	// 猜测原因：长短信头有 6 位协议头(我们平台使用) 和 7位协议头，应该是为了兼容 7 位协议头
)

var (
	ErrUnsupportedDataCoding = errors.New("unsupported data coding")
	ErrInvalidCharacter      = errors.New("invalid character")
)

type DataCoding string

const (
	DataCodingASCII        DataCoding = "ASCII"
	DataCodingGB18030      DataCoding = "GB18030"
	DataCodingGSM7Packed   DataCoding = "GSM 7-bit (Packed)"
	DataCodingGSM7UnPacked DataCoding = "GSM 7-bit (Unpacked)"
	DataCodingLatin1       DataCoding = "LATIN1"
	DataCodingUcs2         DataCoding = "UCS2"
)

func (d DataCoding) String() string {
	return string(d)
}

type ProtocolDataCoding interface {
	ToInt() int

	ToUint8() uint8

	// String returns the name of ProtocolDataCoding.
	String() string

	// Priority returns the priority of ProtocolDataCoding.
	Priority() int
}

// Codec defines a text datacoding.
type Codec interface {
	// Name returns the name of DataCoding.
	Name() DataCoding

	// Encode text.
	Encode() ([]byte, error)

	// Decode text.
	Decode() ([]byte, error)

	// SplitBy returns the maximum length of the currently coded text message,
	// and how much it should be split when it exceeds this length
	SplitBy() (maxLen, splitBy int)
}

// IsValidProtoDataCoding ...
func IsValidProtoDataCoding(dataCoding ProtocolDataCoding) bool {
	if dataCoding == nil {
		return false
	}

	switch t := dataCoding.(type) {
	case CMPPDataCoding:
		return IsValidCMPPDataCoding(t)
	case SMPPDataCoding:
		return IsValidSMPPDataCoding(t)
	}

	return false
}

var UnknownProtocolDataCoding ProtocolDataCoding = unknownProtocolDataCoding(0)

type unknownProtocolDataCoding int

func (u unknownProtocolDataCoding) ToInt() int {
	return int(u)
}

func (u unknownProtocolDataCoding) ToUint8() uint8 {
	return uint8(u)
}

func (u unknownProtocolDataCoding) String() string {
	return "UNKNOWN_DATA_CODING"
}

func (u unknownProtocolDataCoding) Priority() int {
	return math.MinInt
}
