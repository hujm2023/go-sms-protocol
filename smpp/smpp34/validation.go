package smpp34

import "github.com/hujm2023/go-sms-protocol/smpp"

func validateMessageAddressFields(serviceType string, sourceTON, sourceNPI uint8, sourceAddress string, destTON, destNPI uint8, destinationAddress string) error {
	if err := smpp.ValidateCString(smpp.SERVICE_TYPE, serviceType, 6); err != nil {
		return err
	}
	if err := smpp.ValidateTON(smpp.SOURCE_ADDR_TON, sourceTON); err != nil {
		return err
	}
	if err := smpp.ValidateNPI(smpp.SOURCE_ADDR_NPI, sourceNPI); err != nil {
		return err
	}
	if err := smpp.ValidateCString(smpp.SOURCE_ADDR, sourceAddress, 21); err != nil {
		return err
	}
	if err := smpp.ValidateTON(smpp.DEST_ADDR_TON, destTON); err != nil {
		return err
	}
	if err := smpp.ValidateNPI(smpp.DEST_ADDR_NPI, destNPI); err != nil {
		return err
	}
	return smpp.ValidateCString(smpp.DESTINATION_ADDR, destinationAddress, 21)
}
