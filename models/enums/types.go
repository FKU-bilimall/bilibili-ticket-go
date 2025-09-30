package enums

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
