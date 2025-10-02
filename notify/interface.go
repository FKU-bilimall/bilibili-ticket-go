package notify

import (
	"bilibili-ticket-go/global"
	"bilibili-ticket-go/utils"
)

var logger = utils.GetLogger(global.GetLogger(), "notify", nil)

type Notify interface {
	Notify(message string) bool
	Test() bool
}
