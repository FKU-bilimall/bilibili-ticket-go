package enums

import "strings"

type CaptchaType int

const (
	Slide CaptchaType = iota
	Click
	Unknown
)

func (c CaptchaType) String() string {
	switch c {
	case Slide:
		return "Slide"
	case Click:
		return "Click"
	default:
		return "Unknown"
	}
}

type BuyerType int

const (
	Ordinary BuyerType = iota + 1
	ForceRealName
)

type NotificationType int

const (
	None NotificationType = iota
	Gotify
)

func ConvertNotificationType(s string) NotificationType {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "gotify":
		return Gotify
	default:
		return None
	}
}
