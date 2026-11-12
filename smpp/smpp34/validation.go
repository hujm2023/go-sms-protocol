package smpp34

import (
	"fmt"
	"strings"

	"github.com/hujm2023/go-sms-protocol/smpp"
)

func validateMessageAddressFields(serviceType string, sourceTON, sourceNPI uint8, sourceAddress string, destTON, destNPI uint8, destinationAddress string) error {
	if err := smpp.ValidateCString(smpp.SERVICE_TYPE, serviceType, smpp.MAX_PDU_SIZE); err != nil {
		return err
	}
	if err := smpp.ValidateTON(smpp.SOURCE_ADDR_TON, sourceTON); err != nil {
		return err
	}
	if err := smpp.ValidateNPI(smpp.SOURCE_ADDR_NPI, sourceNPI); err != nil {
		return err
	}
	if err := validateAddress(smpp.SOURCE_ADDR, sourceAddress); err != nil {
		return err
	}
	if err := smpp.ValidateTON(smpp.DEST_ADDR_TON, destTON); err != nil {
		return err
	}
	if err := smpp.ValidateNPI(smpp.DEST_ADDR_NPI, destNPI); err != nil {
		return err
	}
	return validateAddress(smpp.DESTINATION_ADDR, destinationAddress)
}

func validateAddress(field, value string) error {
	if len(value) >= smpp.MAX_PDU_SIZE {
		return fmt.Errorf("%s exceeds PDU capacity", field)
	}
	if strings.IndexByte(value, 0) >= 0 {
		return fmt.Errorf("%s must not contain NULL", field)
	}
	return nil
}
