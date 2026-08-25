package smpp

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

var ErrInvalidSequenceNumber = errors.New("invalid SMPP sequence_number")

// ValidateDecodedPDU validates the common SMPP header against the bytes that
// contain the complete PDU. sequence_number is encoded as an unsigned
// 4-octet Integer, but SMPP 3.4 limits its value to 1..0x7fffffff. A zero
// value is only permitted for a generic_nack when the peer could not recover
// the original sequence_number.
func ValidateDecodedPDU(data []byte, expectedID CMDId, allowZeroSequence bool) (Header, error) {
	if len(data) < MinSMPPPacketLen || len(data) > MAX_PDU_SIZE {
		return Header{}, fmt.Errorf("%w: got %d bytes", ErrInvalidPudLength, len(data))
	}

	header, err := PeekHeader(data)
	if err != nil {
		return Header{}, err
	}
	if header.Length < MinSMPPPacketLen || header.Length > MAX_PDU_SIZE || uint64(header.Length) != uint64(len(data)) {
		return Header{}, fmt.Errorf("%w: command_length=%d, actual=%d", ErrInvalidPudLength, header.Length, len(data))
	}
	if header.ID != expectedID {
		return Header{}, fmt.Errorf("unexpected command_id: got %s, want %s", header.ID, expectedID)
	}
	if err := ValidateSequenceNumber(header.Sequence, allowZeroSequence); err != nil {
		return Header{}, err
	}

	return header, nil
}

// ValidateRequestHeader validates fields common to request PDUs before encode.
// Header.Length is intentionally ignored because encoders calculate it from
// the serialized PDU.
func ValidateRequestHeader(header Header, expectedID CMDId) error {
	if err := validateHeaderIdentity(header, expectedID, false); err != nil {
		return err
	}
	if header.Status != ESME_ROK {
		return fmt.Errorf("request command_status must be ESME_ROK, got %s", header.Status)
	}
	return nil
}

// ValidateResponseHeader validates fields common to response PDUs before
// encode. Responses may carry any command_status defined by the peer.
func ValidateResponseHeader(header Header, expectedID CMDId) error {
	return validateHeaderIdentity(header, expectedID, false)
}

// ValidateGenericNackHeader validates the generic_nack exception: the
// sequence_number may be zero and command_status must describe an error.
func ValidateGenericNackHeader(header Header) error {
	if err := validateHeaderIdentity(header, GENERIC_NACK, true); err != nil {
		return err
	}
	if header.Status == ESME_ROK {
		return errors.New("generic_nack command_status must not be ESME_ROK")
	}
	return nil
}

func validateHeaderIdentity(header Header, expectedID CMDId, allowZeroSequence bool) error {
	if header.ID != expectedID {
		return fmt.Errorf("unexpected command_id: got %s, want %s", header.ID, expectedID)
	}
	return ValidateSequenceNumber(header.Sequence, allowZeroSequence)
}

func ValidateSequenceNumber(sequence uint32, allowZero bool) error {
	if allowZero && sequence == 0 {
		return nil
	}
	if sequence < SEQUENCE_NUM_START || sequence > SEQUENCE_NUM_END {
		return fmt.Errorf("%w: %#08x is outside %#08x..%#08x", ErrInvalidSequenceNumber, sequence, SEQUENCE_NUM_START, SEQUENCE_NUM_END)
	}
	return nil
}

// ValidateCString validates an SMPP C-Octet String. maxFieldOctets includes
// the terminating NULL octet, as it does in the SMPP 3.4 field tables.
func ValidateCString(field, value string, maxFieldOctets int) error {
	if maxFieldOctets < 1 {
		return fmt.Errorf("%s has invalid maximum size %d", field, maxFieldOctets)
	}
	if len(value)+1 > maxFieldOctets {
		return fmt.Errorf("%s is %d octets; maximum content is %d", field, len(value), maxFieldOctets-1)
	}
	if !utf8.ValidString(value) {
		return fmt.Errorf("%s is not valid text", field)
	}
	for _, b := range []byte(value) {
		if b == 0 || b > 0x7f {
			return fmt.Errorf("%s must contain ASCII octets other than NULL", field)
		}
	}
	return nil
}

func ValidateTON(field string, value uint8) error {
	if value > TON_Abbreviated {
		return fmt.Errorf("%s has invalid TON value %#02x", field, value)
	}
	return nil
}

func ValidateNPI(field string, value uint8) error {
	switch value {
	case NPI_Unknown, NPI_ISDN, NPI_Data, NPI_Telex, NPI_LandMobile,
		NPI_National, NPI_Private, NPI_ERMES, NPI_Internet, NPI_WAPClientID:
		return nil
	default:
		return fmt.Errorf("%s has invalid NPI value %#02x", field, value)
	}
}

func ValidateInterfaceVersion(value uint8) error {
	if value > 0x34 {
		return fmt.Errorf("interface_version %#02x exceeds SMPP 3.4", value)
	}
	return nil
}

func ValidateDataCoding(value uint8) error {
	if value <= ENCODING_ISO2022JP || value == ENCODING_EXTJIS || value == ENCODING_KSC5601 ||
		(value >= 0xc0 && value <= 0xdf) || value >= 0xf0 {
		return nil
	}
	return fmt.Errorf("data_coding %#02x is reserved by SMPP 3.4", value)
}

func ValidatePriorityFlag(value uint8) error {
	if value > 3 {
		return fmt.Errorf("priority_flag must be in 0..3, got %d", value)
	}
	return nil
}

// ValidateTimeCString validates the 16-character absolute or relative SMPP
// time format. Empty is the one-octet NULL representation.
func ValidateTimeCString(field, value string) error {
	if value == "" {
		return nil
	}
	if len(value) != 16 {
		return fmt.Errorf("%s must be empty or contain exactly 16 ASCII characters", field)
	}
	for i := 0; i < 15; i++ {
		if value[i] < '0' || value[i] > '9' {
			return fmt.Errorf("%s contains a non-digit at offset %d", field, i)
		}
	}

	pair := func(offset int) int {
		return int(value[offset]-'0')*10 + int(value[offset+1]-'0')
	}
	checkMax := func(name string, got, max int) error {
		if got > max {
			return fmt.Errorf("%s %s component %d exceeds %d", field, name, got, max)
		}
		return nil
	}

	switch value[15] {
	case '+', '-':
		if month := pair(2); month < 1 || month > 12 {
			return fmt.Errorf("%s month component %d is outside 1..12", field, month)
		}
		if day := pair(4); day < 1 || day > 31 {
			return fmt.Errorf("%s day component %d is outside 1..31", field, day)
		}
		if err := checkMax("hour", pair(6), 23); err != nil {
			return err
		}
		if err := checkMax("minute", pair(8), 59); err != nil {
			return err
		}
		if err := checkMax("second", pair(10), 59); err != nil {
			return err
		}
		if err := checkMax("UTC offset", pair(13), 48); err != nil {
			return err
		}
	case 'R':
		if err := checkMax("month", pair(2), 11); err != nil {
			return err
		}
		if err := checkMax("day", pair(4), 30); err != nil {
			return err
		}
		if err := checkMax("hour", pair(6), 23); err != nil {
			return err
		}
		if err := checkMax("minute", pair(8), 59); err != nil {
			return err
		}
		if err := checkMax("second", pair(10), 59); err != nil {
			return err
		}
		if value[12] != '0' || pair(13) != 0 {
			return fmt.Errorf("%s relative time tenths and UTC-offset fields must be zero", field)
		}
	default:
		return fmt.Errorf("%s orientation must be '+', '-', or 'R'", field)
	}

	return nil
}

func ValidateEncodedPDU(data []byte) error {
	if len(data) < MinSMPPPacketLen || len(data) > MAX_PDU_SIZE {
		return fmt.Errorf("%w: encoded PDU is %d bytes", ErrInvalidPudLength, len(data))
	}
	return nil
}
