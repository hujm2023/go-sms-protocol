package datacoding

import (
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// UCS2 text datacoding.
type UCS2 []byte

// Name implements the Codec interface.
func (s UCS2) Name() DataCoding {
	return DataCodingUcs2
}

// Encode to UCS2.
func (s UCS2) Encode() ([]byte, error) {
	e := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
	es, _, err := transform.Bytes(e.NewEncoder(), s)
	if err != nil {
		return s, err
	}
	return es, nil
}

// Decode from UCS2.
func (s UCS2) Decode() ([]byte, error) {
	e := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
	es, _, err := transform.Bytes(e.NewDecoder(), s)
	if err != nil {
		return s, err
	}
	return es, nil
}

func (s UCS2) SplitBy() (maxLen, splitBy int) {
	return MaxLongSmsLength, SplitBy134
}
