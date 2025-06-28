package desktop

import (
	"bytes"
	"encoding/json"
	"fmt"
	"popmart/src/middleware/helpers"
	"popmart/src/middleware/helpers/adyen"
)

func (om OrderedMap) MarshalJSON() ([]byte, error) {
	return MarshalOrderedMap([]OrderedKV(om))
}

func MarshalOrderedMap(pairs []OrderedKV) (json.RawMessage, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, pair := range pairs {
		keyJSON, _ := json.Marshal(pair.Key)
		valJSON, _ := json.Marshal(pair.Value)

		buf.Write(keyJSON)
		buf.WriteByte(':')
		buf.Write(valJSON)

		if i < len(pairs)-1 {
			buf.WriteByte(',')
		}
	}
	buf.WriteByte('}')
	return json.RawMessage(buf.Bytes()), nil
}

func AdyenHelper(logger *helpers.ColorizedLogger, task helpers.Task, order OrderDetails, checkoutAttemptId string) (string, error) {
	adyenData, err := adyen.AdyenEncrypt(logger, task.TaskId, task.Profile.CardNumber, task.Profile.ExpMonth, task.Profile.ExpYear, task.Profile.CVV)
	if err != nil {
		logger.Error(fmt.Sprintf("Task %s: Failed To Adyen Encrypt Payment", task.TaskId))
		return "", err
	}

	adyenPayload := AdyenData{
		PaymentMethod: AdyenPayment{
			Type:              "scheme",
			HolderName:        task.Profile.Name,
			CardNumber:        adyenData.CardNumber,
			ExpiryMonth:       adyenData.ExpiryMonth,
			ExpiryYear:        adyenData.ExpiryYear,
			SecurityCode:      adyenData.SecurityCode,
			CardBrand:         adyenData.CardType,
			CheckoutAttemptId: checkoutAttemptId,
		},
		BrowserInfo: AdyenBrowser{
			TimezoneOffset:    240,
			AcceptHeader:      "*/*",
			JavascriptEnabled: true,
			Language:          "nl-NL",
			JavaEnabled:       true,
			ScreenHeight:      723,
			ScreenWidth:       1536,
			ColorDepth:        24,
			UserAgent:         helpers.UserAgent,
		},
		StorePayment: false,
		Risk: RiskData{
			ClientData: adyenData.RiskData,
		},
		AdditionalData: Additional{
			Allow3DS2: "true",
		},
		Channel:   "Web",
		Origin:    "https://www.popmart.com",
		ReturnUrl: fmt.Sprintf("https://www.popmart.com/us/checkout?type=normal&orderNo=%s&payType=adyen", order.OrderNumber),
	}

	payloadBytes, err := json.Marshal(adyenPayload)
	if err != nil {
		logger.Error(fmt.Sprintf("Task %s: Failed to Marshal Adyen Payload", task.TaskId))
		return "", err
	}
	return string(payloadBytes), nil
}
