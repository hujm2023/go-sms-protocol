package datacoding

import (
	"unicode/utf8"
)

type Ascii string

func (a Ascii) Name() DataCoding {
	return DataCodingASCII
}

func (a Ascii) Encode() ([]byte, error) {
	if !isASCII(string(a)) {
		return nil, ErrInvalidCharacter
	}
	return []byte(a), nil
}

func (a Ascii) Decode() ([]byte, error) {
	if !isASCII(string(a)) {
		return nil, ErrInvalidCharacter
	}
	return []byte(a), nil
}

func (a Ascii) SplitBy() (maxLen, splitBy int) {
	return MaxLongSmsLength, SplitBy134
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}
