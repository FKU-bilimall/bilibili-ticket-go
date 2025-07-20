package hooks

import (
	"bilibili-ticket-go/utils"
	"github.com/DeRuina/timberjack"
	"github.com/sirupsen/logrus"
	"sync"
)

func NewLogFileRotateHook(logger *timberjack.Logger) logrus.Hook {
	return &LogFileRotateHook{logger: logger, mutex: &sync.Mutex{}}
}

type LogFileRotateHook struct {
	logger *timberjack.Logger
	mutex  *sync.Mutex
}

func (h *LogFileRotateHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *LogFileRotateHook) Fire(entry *logrus.Entry) error {
	msg, err := entry.String()
	msg = utils.ANSIStrip(msg)
	if err != nil {
		return err
	}
	h.mutex.Lock()
	defer h.mutex.Unlock()
	_, err = h.logger.Write([]byte(msg))
	return err
}
