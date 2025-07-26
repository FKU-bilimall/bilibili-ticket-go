package api

import (
	"errors"
	"fmt"
)

// MainApiDataRoot 主站API基类
type MainApiDataRoot[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func (r *MainApiDataRoot[T]) CheckValid() error {
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

type WbiStruct struct {
	WbiImg struct {
		ImgUrl string `json:"img_url"`
		SubUrl string `json:"sub_url"`
	} `json:"wbi_img"`
}

type TicketProjectInformationStruct struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	SaleBegin  int    `json:"sale_begin"`
	SaleEnd    int    `json:"sale_end"`
	HotProject bool   `json:"hotProject"`
	ScreenList []struct {
		SaleFlag struct {
			Number      int    `json:"number"`
			DisplayName string `json:"display_name"`
		} `json:"saleFlag"`
		ScreenId     int    `json:"id"`
		StartTime    int    `json:"start_time"`
		Name         string `json:"name"`
		Type         int    `json:"type"`
		TicketType   int    `json:"ticket_type"`
		ScreenType   int    `json:"screen_type"`
		DeliveryType int    `json:"delivery_type"`
		PickSeat     int    `json:"pick_seat"`
		TicketList   []struct {
			Price     int    `json:"price"`
			Desc      string `json:"desc"`
			SaleStart int    `json:"saleStart"`
			SaleEnd   int    `json:"saleEnd"`
			IsSale    int    `json:"is_sale"`
			SkuId     int    `json:"id"`
			SaleFlag  struct {
				Number      int    `json:"number"`
				DisplayName string `json:"display_name"`
			} `json:"sale_flag"`
			ScreenName string `json:"screen_name"`
		} `json:"ticket_list"`
	} `json:"screen_list"`
}
