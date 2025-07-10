package response

type Root[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    T      `json:"data"`
}

type GetQRLoginKeyPayload struct {
	URL       string `json:"url"`
	QRCodeKey string `json:"qrcode_key"`
}

type VerifyQRLoginStatePayload struct {
	RefreshToken string `json:"refresh_token"`
	Timestamp    int64  `json:"timestamp"`
	Code         int    `json:"code"`
	Message      string `json:"message"`
}

type GetLoginInfoPayload struct {
	Login bool   `json:"isLogin"`
	Name  string `json:"uname,omitempty"`
	UID   int64  `json:"mid,omitempty"`
}

type GetBVUID34Payload struct {
	BVUID3 string `json:"b_3"`
	BVUID4 string `json:"b_4"`
}

type NeedRefreshPayload struct {
	NeedRefresh bool  `json:"refresh"`
	Timestamp   int64 `json:"timestamp"`
}
