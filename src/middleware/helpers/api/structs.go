package api

type TdPayload struct {
	TaskId    string         `json:"taskId"`
	Path      string         `json:"path"`
	Payload   map[string]any `json:"payload"`
	ProxyUrl  string         `json:"proxyUrl"`
	Method    string         `json:"method"`
	UserAgent string         `json:"userAgent"`
}

type TdResp struct {
	Code      string `json:"code"`
	Result    string `json:"result"`
	RequestId string `json:"requestId"`
}

type DecodeResp struct {
	Bxid    string     `json:"bxid"`
	C       CryptoMeta `json:"c"`
	TokenId string     `json:"tokenId"`
	Xdid    string     `json:"xdid"`
	Xxid    string     `json:"xxid"`
}

type CryptoMeta struct {
	Cm     int `json:"cm"`
	Factor int `json:"factor"`
	Op     int `json:"op"`
	Vt     int `json:"vt"`
}

type SessionResp struct {
	SessionSign string `json:"td-session-sign"`
	SessionKey  string `json:"td-session-key"`
}

type ApiResp struct {
	Success     bool   `json:"success"`
	S           string `json:"s"`
	T           int    `json:"t"`
	Sign        string `json:"sign"`
	SessionSign string `json:"sessionSign"`
	SessionKey  string `json:"sessionKey"`
}
