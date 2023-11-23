package datacoding

import (
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

// Latin1 text datacoding.
type Latin1 []byte

// Name implements the Codec interface.
func (s Latin1) Name() DataCoding {
	return DataCodingLatin1
}

// Encode to Latin1.
func (s Latin1) Encode() ([]byte, error) {
	e := charmap.Windows1252.NewEncoder()
	es, _, err := transform.Bytes(e, s)
	if err != nil {
		return s, err
	}
	return es, nil
}

// Decode from Latin1.
func (s Latin1) Decode() ([]byte, error) {
	e := charmap.Windows1252.NewDecoder()
	es, _, err := transform.Bytes(e, s)
	if err != nil {
		return s, err
	}
	return es, nil
}

func (s Latin1) SplitBy() (maxLen, splitBy int) {
	return MaxLongSmsLength, SplitBy134
}
