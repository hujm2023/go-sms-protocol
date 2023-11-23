package datacoding

import (
	encoding "github.com/hujm2023/go-sms-protocol/datacoding/gsm7encoding"
)

// GSM7Packed GSM 7-bit (packed)
type GSM7Packed []byte

// Name implements the Codec interface.
func (s GSM7Packed) Name() DataCoding {
	return DataCodingGSM7Packed
}

// Encode to GSM 7-bit (packed)
func (s GSM7Packed) Encode() ([]byte, error) {
	src, err := encoding.Encode(string(s))
	if err != nil {
		return nil, err
	}
	return encoding.Pack(src), nil
	// e := encoding.GSM7(true).NewEncoder()
	// es, _, err := transform.Bytes(e, s)
	// if err != nil {
	// 	return s, err
	// }
	// return es, nil
}

// Decode from GSM 7-bit (packed)
func (s GSM7Packed) Decode() ([]byte, error) {
	return encoding.Decode(encoding.Unpack(s))
	// e := encoding.GSM7(true).NewDecoder()
	// es, _, err := transform.Bytes(e, s)
	// if err != nil {
	// 	return s, err
	// }
	// return es, nil
}

func (s GSM7Packed) SplitBy() (maxLen, splitBy int) {
	return MaxGSM7Length, SplitBy153
}
