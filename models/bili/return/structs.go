package _return

import (
	"bilibili-ticket-go/models/enums"
	"strconv"
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
	ProjectID       string
	StartTime       time.Time
	EndTime         time.Time
	IsHotProject    bool
	IsForceRealName bool
	IsNeedContact   bool
	ProjectName     string
}

type TicketBuyer struct {
	BuyerType enums.BuyerType
	ID        int64
	Tel       string
	Name      string
}

func (buyer TicketBuyer) Compare(a TicketBuyer) bool {
	if buyer.BuyerType != a.BuyerType {
		return false
	}
	if buyer.BuyerType == enums.Ordinary {
		return buyer.Tel == a.Tel && buyer.Name == a.Name
	} else {
		return buyer.ID == a.ID
	}
}

func (buyer TicketBuyer) String() string {
	if buyer.BuyerType == enums.Ordinary {
		return buyer.Name + " (" + buyer.Tel + ")"
	} else {
		return buyer.Name + " (ID: " + strconv.FormatInt(buyer.ID, 10) + ")"
	}
}
