package datacoding

// cmppDataCodingPriority 编码后内容长度优先级。数值越小，优先级越高
var cmppDataCodingPriority map[CMPPDataCoding]int

func init() {
	cmppDataCodingPriority = make(map[CMPPDataCoding]int)

	cmppDataCodingPriority[CMPP_CODING_UCS2] = 2 << 0
	cmppDataCodingPriority[CMPP_CODING_GBK] = 2 << 1
	cmppDataCodingPriority[CMPP_CODING_ASCII] = 2 << 2
}

// CMPPDataCoding represents the encoding and decoding types supported by the cmpp protocol.
type CMPPDataCoding int

const (
	CMPP_CODING_ASCII CMPPDataCoding = 0  // ascii
	CMPP_CODING_UCS2  CMPPDataCoding = 8  // ucs2
	CMPP_CODING_GBK   CMPPDataCoding = 15 // gbk
)

// ToUint8 返回协议中对应编码的枚举值
func (c CMPPDataCoding) ToUint8() uint8 {
	switch c {
	case CMPP_CODING_ASCII:
		return 0
	case CMPP_CODING_UCS2:
		return 8
	case CMPP_CODING_GBK:
		return 15
	default:
		return 255
	}
}

func (c CMPPDataCoding) ToInt() int {
	return int(c.ToUint8())
}

func (c CMPPDataCoding) String() string {
	switch c {
	case CMPP_CODING_ASCII:
		return "ASCII"
	case CMPP_CODING_UCS2:
		return "UCS2"
	case CMPP_CODING_GBK:
		return "GBK"
	default:
		return "UNKNOWN"
	}
}

func (c CMPPDataCoding) Priority() int {
	return cmppDataCodingPriority[c]
}

// GetCMPPCodec creates a Codec with content based on the CMPPDataCoding.
// Note that when there is an unsupported CMPPDataCoding, UCS2 is used as the default.
func GetCMPPCodec(dataCoding CMPPDataCoding, content string) Codec {
	c := NewCMPPCodec(dataCoding, content)
	if c == nil {
		c = UCS2(content) // ucs2兜底
	}
	return c
}

func NewCMPPCodec(dataCoding CMPPDataCoding, content string) Codec {
	switch dataCoding {
	case CMPP_CODING_ASCII:
		return Ascii(content)
	case CMPP_CODING_GBK:
		return GB18030(content)
	case CMPP_CODING_UCS2:
		return UCS2(content)
	default:
		return nil
	}
}

// IsValidCMPPDataCoding returns true if dataCoding is a valid CMPPDataCoding.
func IsValidCMPPDataCoding(dataCoding CMPPDataCoding) bool {
	_, ok := cmppDataCodingPriority[dataCoding]
	return ok
}
