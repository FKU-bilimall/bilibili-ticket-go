package global

import (
	"bilibili-ticket-go/utils"

	"github.com/sirupsen/logrus"
)

var loggers = map[string]*logrus.Logger{
	"main": logrus.New(),
}

func init() {
	utils.RegisterLoggerFormater(loggers["main"])
	loggers["main"].SetLevel(logrus.TraceLevel)
}

func GetLogger() *logrus.Logger {
	return loggers["main"]
}
