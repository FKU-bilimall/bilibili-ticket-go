package _return

import (
	"bilibili-ticket-go/models/bili/api"
	"time"
)

type TicketSkuScreenID struct {
	ScreenID int64
	SkuID    int64
	Name     string
	Desc     string
	Price    int
	Flags    struct {
		Number      int
		DisplayName string
	}
	SaleStat struct {
		Start time.Time
		End   time.Time
	}
}

type RequestTokenAndPToken struct {
	RequestToken string
	PToken       string
	GaiaToken    string
}

type ProjectInformation struct {
	ProjectID    string
	StartTime    time.Time
	EndTime      time.Time
	IsHotProject bool
}

type BuyerInformation struct {
	ForceRealNameBuyer *api.BuyerStruct
	ContactInfo        *ContactInfoStruct
}

type ContactInfoStruct struct {
	Tel      string `json:"tel"`
	Uid      int64  `json:"uid"`
	Username string `json:"username"`
}
