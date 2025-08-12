package _return

import "time"

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
