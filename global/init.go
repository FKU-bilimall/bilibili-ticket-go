package global

import (
	"bilibili-ticket-go/utils"
	"strconv"

	"github.com/sirupsen/logrus"
)

var loggers = map[string]*logrus.Logger{
	"main": logrus.New(),
}

func init() {
	level, err := strconv.Atoi(LoggerLevel)
	if err != nil {
		level = 4
	}
	for _, logger := range loggers {
		logger.SetLevel(logrus.Level(level))
	}
	utils.RegisterLoggerFormater(loggers["main"])
}

func GetLogger() *logrus.Logger {
	return loggers["main"]
}
