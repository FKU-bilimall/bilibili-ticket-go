package _return

import "time"

type TicketSkuScreenID struct {
	ScreenID int
	SkuID    int
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
