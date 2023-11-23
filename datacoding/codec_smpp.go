package datacoding

var smppDataCodingPriority map[SMPPDataCoding]int

func init() {
	smppDataCodingPriority = make(map[SMPPDataCoding]int)

	smppDataCodingPriority[SMPP_CODING_UCS2] = 2 << 0
	smppDataCodingPriority[SMPP_CODING_GSM7_UNPACKED] = 2 << 1
	smppDataCodingPriority[SMPP_CODING_Latin1] = 2 << 2
	smppDataCodingPriority[SMPP_CODING_ASCII] = 2 << 3
	smppDataCodingPriority[SMPP_CODING_GSM7_PACKED] = 2 << 4
}

// SMPPDataCoding represents the encoding and decoding types supported by the smpp protocol.
type SMPPDataCoding int

const (
	SMPP_CODING_GSM7_UNPACKED SMPPDataCoding = 0  // gsm7_unpacked
	SMPP_CODING_GSM7_PACKED   SMPPDataCoding = 99 // gsm7_packed(非协议值，自定义)
	SMPP_CODING_ASCII         SMPPDataCoding = 1  // ascii
	SMPP_CODING_Latin1        SMPPDataCoding = 3  // latin1
	SMPP_CODING_UCS2          SMPPDataCoding = 8  // ucs2
)

// ToInt 返回协议中对应编码的枚举值
func (s SMPPDataCoding) ToInt() int {
	return int(s.ToUint8())
}

func (s SMPPDataCoding) ToUint8() uint8 {
	switch s {
	case SMPP_CODING_GSM7_PACKED, SMPP_CODING_GSM7_UNPACKED:
		return 0x00
	case SMPP_CODING_ASCII:
		return 0x01
	case SMPP_CODING_Latin1:
		return 0x03
	case SMPP_CODING_UCS2:
		return 0x08
	default:
		return 255
	}
}

// String 编码名称。注意不要包含 metrics 不能 emit 的字符，比如空格、括号等
func (s SMPPDataCoding) String() string {
	switch s {
	case SMPP_CODING_GSM7_PACKED:
		return "GSM7_PACKED"
	case SMPP_CODING_GSM7_UNPACKED:
		return "GSM7_UNPACKED"
	case SMPP_CODING_ASCII:
		return "ASCII"
	case SMPP_CODING_Latin1:
		return "Latin1"
	case SMPP_CODING_UCS2:
		return "UCS2"
	default:
		return "UNKNOWN"
	}
}

func (s SMPPDataCoding) Priority() int {
	return smppDataCodingPriority[s]
}

// GetSMPPCodec creates a Codec with content based on the SMPPDataCoding.
// Note that when there is an unsupported SMPPDataCoding, UCS2 is used as the default.
func GetSMPPCodec(dataCoding SMPPDataCoding, content string) Codec {
	c := NewSMPPCodec(dataCoding, content)
	if c == nil {
		c = UCS2(content)
	}
	return c
}

// NewSMPPCodec ...
func NewSMPPCodec(dataCoding SMPPDataCoding, content string) Codec {
	switch dataCoding {
	case SMPP_CODING_GSM7_PACKED:
		return GSM7Packed(content)
	case SMPP_CODING_GSM7_UNPACKED:
		return GSM7Unpacked(content)
	case SMPP_CODING_ASCII:
		return Ascii(content)
	case SMPP_CODING_Latin1: // 默认编码为 Latin1
		return Latin1(content)
	case SMPP_CODING_UCS2:
		return UCS2(content)
	default:
		return nil
	}
}

// IsValidSMPPDataCoding returns true if dataCoding is a valid SMPPDataCoding.
func IsValidSMPPDataCoding(dataCoding SMPPDataCoding) bool {
	_, ok := smppDataCodingPriority[dataCoding]
	return ok
}
