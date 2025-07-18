package response

import (
	"errors"
	"fmt"
)

type DataRoot[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func (r *DataRoot[T]) CheckValid() error {
	if r.Code != 0 {
		return errors.New(fmt.Sprintf("Response code is not 0, got: %d, message: %s", r.Code, r.Message))
	}
	return nil
}

type GetQRLoginKeyStruct struct {
	URL       string `json:"url"`
	QRCodeKey string `json:"qrcode_key"`
}

type VerifyQRLoginStateStruct struct {
	RefreshToken string `json:"refresh_token"`
	Timestamp    int64  `json:"timestamp"`
	Code         int    `json:"code"`
	Message      string `json:"message"`
}

type GetLoginInfoStruct struct {
	Login bool   `json:"isLogin"`
	Name  string `json:"uname,omitempty"`
	UID   int64  `json:"mid,omitempty"`
}

type GetBVUID34Struct struct {
	BVUID3 string `json:"b_3"`
	BVUID4 string `json:"b_4"`
}

type NeedRefreshStruct struct {
	NeedRefresh bool  `json:"refresh"`
	Timestamp   int64 `json:"timestamp"`
}

type RefreshTokenStruct struct {
	RefreshToken string `json:"refresh_token"`
}

type BiliTicketStruct struct {
	Ticket  string `json:"ticket"`
	Created int    `json:"create_at"`
	TTL     int    `json:"ttl"`
}

type BiliAppVersionStruct struct {
	Version string `json:"version"`
	Build   int    `json:"build"`
}
