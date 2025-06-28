package adyen

import (
	"fmt"

	helpers "popmart/src/middleware/helpers"
)

func AdyenEncrypt(logger *helpers.ColorizedLogger, taskId, card, month, year, cvc string) (AdyenResp, error) {
	logger.Info(fmt.Sprintf("Task %s: Adyen Encrypting Payment Information", taskId))
	if card == "" || month == "" || year == "" || cvc == "" {
		logger.Error(fmt.Sprintf("Task %s: Missing Required Encryption Parameters", taskId))
		return AdyenResp{}, fmt.Errorf("missing required parameters")
	}

	enc, err := PrepareEncryptor(adyenKey, liveKey, "https://prod-na-app.popmart.com/shop/v1/shop/cash/desk/adyen/pay")
	if err != nil {
		logger.Error(fmt.Sprintf("Task %s: Failed To Initialize Adyen Encryptor", taskId))
		return AdyenResp{}, fmt.Errorf("failed to initialize encryptor")
	}

	encryptedData, err := enc.EncryptData(card, month, year, cvc)
	if err != nil {
		logger.Error(fmt.Sprintf("Task %s: Failed To Adyen Encrypt Payment Information", taskId))
		return AdyenResp{}, fmt.Errorf("failed to initialize encryptor")
	}

	rd := NewRiskData(
		helpers.UserAgent,
		"en-US", 24, 4, 8, 360, 640, 360, 640, -300,
		"America/Chicago", "Windows", nil, nil,
	)

	var cardType string
	if len(card) > 0 {
		switch card[0] {
		case '4':
			cardType = "visa"
		case '5':
			cardType = "mastercard"
		case '3':
			cardType = "amex"
		case '6':
			cardType = "discover"
		default:
			cardType = "unknown"
		}
	}

	response := AdyenResp{
		Success:      true,
		RiskData:     rd.Generate(),
		CardNumber:   encryptedData.EncryptedCardNumber,
		ExpiryMonth:  encryptedData.EncryptedExpiryMonth,
		ExpiryYear:   encryptedData.EncryptedExpiryYear,
		SecurityCode: encryptedData.EncryptedSecurityCode,
		CardType:     cardType,
	}
	return response, nil
}
