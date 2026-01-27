package helpers

import (
	"billpayment_service/models"
	"encoding/base64"
)

func GetNetworkCode(networkName string, serviceType string) (resp string) {
	networkCode := networkName + "_" + serviceType

	return networkCode
}

func GetServiceId(network string) (string, error) {

	if networkService, err := models.GetNetworksByCode(network); err == nil {
		return networkService.NetworkReferenceId, nil
	}

	return "", nil
}

func ConvertToBase64(input string) string {
	encoded := ""
	// encoding logic here
	encoded = base64.StdEncoding.EncodeToString([]byte(input))
	return encoded
}

func ConvertFromBase64(encoded string) (string, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(decodedBytes), nil
}
