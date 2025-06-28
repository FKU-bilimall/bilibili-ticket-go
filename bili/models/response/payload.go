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
	Timestamp    int    `json:"timestamp"`
	Code         int    `json:"code"`
	Message      string `json:"message"`
}
