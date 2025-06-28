package adyen

import "crypto/rsa"

type Encryptor struct {
	Key       string
	OriginKey string
	Domain    string
	RsaPubKey *rsa.PublicKey
}

type AdyenData struct {
	EncryptedCardNumber   string `json:"encryptedCardNumber"`
	EncryptedExpiryMonth  string `json:"encryptedExpiryMonth"`
	EncryptedExpiryYear   string `json:"encryptedExpiryYear"`
	EncryptedSecurityCode string `json:"encryptedSecurityCode"`
}

type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
	Alg string `json:"alg"`
	Use string `json:"use"`
}

type RiskData struct {
	UserAgent           string
	Language            string
	ColorDepth          int
	DeviceMemory        int
	HardwareConcurrency int
	ScreenWidth         int
	ScreenHeight        int
	AvailScreenWidth    int
	AvailScreenHeight   int
	TimezoneOffset      int
	Timezone            string
	Platform            string
	CpuClass            *string
	DoNotTrack          *string
}

type AdyenResp struct {
	Success      bool   `json:"success"`
	RiskData     string `json:"riskData"`
	CardNumber   string `json:"cardNumber"`
	ExpiryMonth  string `json:"expiryMonth"`
	ExpiryYear   string `json:"expiryYear"`
	SecurityCode string `json:"securityCode"`
	CardType     string `json:"cardType"`
}
