package api

import (
	"bilibili-ticket-go/models/errors"
)

// ShowApiDataRoot 漫展API基类
type ShowApiDataRoot[T any] struct {
	ErrTag    int    `json:"errtag"`
	ErrNumber int    `json:"errno"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Msg       string `json:"msg"`
	Data      T      `json:"data"`
}

func (r *ShowApiDataRoot[T]) GetCode() int {
	if r.ErrNumber != 0 {
		return r.ErrNumber
	} else {
		return r.Code
	}
}

func (r *ShowApiDataRoot[T]) GetMessage() string {
	if r.ErrNumber != 0 {
		return r.Msg
	} else {
		return r.Message
	}
}

func (r *ShowApiDataRoot[T]) CheckValid() error {
	if r.ErrNumber != 0 || r.Code != 0 {
		return errors.NewBilibiliAPIError(r.GetCode(), r.GetMessage())
	}
	return nil
}

type RequestTokenAndPTokenStruct struct {
	Token  string `json:"token"`
	Shield struct {
		Open int `json:"open"`
	} `json:"shield"`
	ProjectName interface{} `json:"project_name"`
	ScreenName  interface{} `json:"screen_name"`
	ProjectImg  interface{} `json:"project_img"`
	GaData      struct {
		RiskLevel  int           `json:"risk_level"`
		GriskId    string        `json:"grisk_id"`
		Decisions  []interface{} `json:"decisions"`
		RiskParams interface{}   `json:"riskParams"`
		RiskResult int           `json:"riskResult"`
		Open       interface{}   `json:"open"`
	} `json:"ga_data"`
	SuccessSeats interface{}   `json:"success_seats"`
	FailedSeats  []interface{} `json:"failed_seats"`
	Ptoken       string        `json:"ptoken"`
}

type ConfirmStruct struct {
	Count     int `json:"count"`
	BuyerList struct {
		List     []BuyerStruct `json:"list"`
		MaxLimit int           `json:"max_limit"`
	} `json:"buyerList"`
	HotProject     bool   `json:"hotProject"`
	OrderCreateUrl string `json:"orderCreateUrl"`
	ProjectId      int    `json:"project_id"`
	ProjectName    string `json:"project_name"`
	ScreenId       int    `json:"screen_id"`
	ScreenName     string `json:"screen_name"`
	BuyerInfo      string `json:"buyer_info"`
	ItemTotalMoney int    `json:"item_total_money"` // its value is the total price of all tickets, often equal to `pay_money`
	PayMoney       int    `json:"pay_money"`
	TicketInfo     struct {
		Name  string `json:"name"`
		Price int    `json:"price"`
		Count int    `json:"count"`
		SkuId int    `json:"sku_id"`
	} `json:"ticket_info"`
}

type BuyerStruct struct {
	Id                  int64       `json:"id"`
	Uid                 int64       `json:"uid"`
	AccountId           int         `json:"accountId"`
	Name                string      `json:"name"`
	Buyer               interface{} `json:"buyer,"`
	Tel                 string      `json:"tel"`
	DisabledErr         interface{} `json:"disabledErr"`
	AccountChannel      string      `json:"account_channel"`
	PersonalId          string      `json:"personal_id"`
	IdCardFront         string      `json:"id_card_front"`
	IdCardBack          string      `json:"id_card_back"`
	IsDefault           int         `json:"is_default"`
	IdType              int         `json:"id_type"`
	VerifyStatus        int         `json:"verify_status"`
	IsBuyerInfoVerified bool        `json:"isBuyerInfoVerified"`
	IsBuyerValid        bool        `json:"isBuyerValid"`
}

type TicketOrderStruct struct {
	OrderId         int64  `json:"orderId"`
	OrderCreateTime int64  `json:"orderCreateTime"`
	Token           string `json:"token"`
	PayMoney        int    `json:"pay_money"`
}

type BuyerNoSensitiveInfoApiStruct struct {
	Vo struct {
		List []BuyerNoSensitiveStruct `json:"list"`
	} `json:"vo"`
}

type BuyerNoSensitiveStruct struct {
	Id           int64  `json:"id"`
	Uid          int64  `json:"uid"`
	Name         string `json:"name"`
	IdType       int    `json:"idType"`
	IdName       string `json:"idName"`
	IdCard       string `json:"idCard"`
	Tel          string `json:"tel"`
	ViewType     string `json:"viewType"`
	VerifyStatus int    `json:"verifyStatus"`
	Status       int    `json:"status"`
}
