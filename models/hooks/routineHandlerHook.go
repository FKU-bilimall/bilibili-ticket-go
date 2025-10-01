package hooks

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type RoutineHandlerHook struct {
	callback func(int, logrus.Fields)
	sync.Mutex
}

func NewRoutineHandlerHook(callback func(int, logrus.Fields)) *RoutineHandlerHook {
	return &RoutineHandlerHook{
		callback: callback,
	}
}

func (h *RoutineHandlerHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *RoutineHandlerHook) Fire(entry *logrus.Entry) error {
	h.Mutex.Lock()
	defer h.Mutex.Unlock()
	stat := entry.Data["status"]
	if stat == nil {
		return nil
	}
	val, ok := stat.(int)
	if !ok {
		return nil
	}
	bili, ok := entry.Data["bili"].(logrus.Fields)
	if !ok {
		bili = logrus.Fields{}
	}
	if h.callback != nil {
		h.callback(val, bili)
	}
	return nil
}
