package utils

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

type LogAdditionalParts struct {
}

func GetLogger(name string, parts *LogAdditionalParts) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"name":  name,
		"parts": parts,
	})
}

func RegisterLoggerFormater() {
	logrus.SetFormatter(&ColorfulFormatter{})
}

type ColorfulFormatter struct {
}

func (f *ColorfulFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var levelColor *color.Color
	switch entry.Level {
	case logrus.DebugLevel:
		levelColor, _ = hexToColor("4caf50")
	case logrus.InfoLevel:
		levelColor, _ = hexToColor("2196f3")
	case logrus.WarnLevel:
		levelColor, _ = hexToColor("ffeb3b")
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor, _ = hexToColor("f44336")
	default:
		levelColor, _ = hexToColor("ffffff")
	}
	return []byte(
		fmt.Sprintf(
			"%s | %s | %s | %s\n",
			entry.Time.Format(time.RFC3339),
			levelColor.Sprint(strings.ToUpper(entry.Level.String()[0:4])),
			strings.ToLower(entry.Data["name"].(string)),
			entry.Message,
		),
	), nil
}
