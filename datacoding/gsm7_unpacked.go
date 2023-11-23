package datacoding

import (
	"golang.org/x/text/transform"

	encoding "github.com/hujm2023/go-sms-protocol/datacoding/gsm7encoding"
)

// GSM7Unpacked GSM 7-bit (unpacked)
type GSM7Unpacked []byte

// Name implements the Codec interface.
func (s GSM7Unpacked) Name() DataCoding {
	return DataCodingGSM7UnPacked
}

// Encode to GSM 7-bit (unpacked)
func (s GSM7Unpacked) Encode() ([]byte, error) {
	e := encoding.GSM7(false).NewEncoder()
	es, _, err := transform.Bytes(e, s)
	if err != nil {
		return s, err
	}
	return es, nil
}

// Decode from GSM 7-bit (unpacked)
func (s GSM7Unpacked) Decode() ([]byte, error) {
	e := encoding.GSM7(false).NewDecoder()
	es, _, err := transform.Bytes(e, s)
	if err != nil {
		return s, err
	}
	return es, nil
}

func (s GSM7Unpacked) SplitBy() (maxLen, splitBy int) {
	return MaxGSM7Length, SplitBy153
}

func CanEncodeByGSM7(text string) bool {
	return encoding.IsValidGSM7String(text)
}
