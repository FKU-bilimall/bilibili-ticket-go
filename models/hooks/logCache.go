package hooks

import (
	"io"
	"sync"

	"github.com/sirupsen/logrus"
)

// LoggerCache 带缓存的日志记录器
type LoggerCache struct {
	mu       sync.RWMutex
	entries  []string
	maxLines int
	out      io.Writer
}

func NewLoggerCache(maxLines int, out io.Writer) *LoggerCache {
	return &LoggerCache{
		entries:  make([]string, 0, maxLines),
		maxLines: maxLines,
		out:      out,
	}
}

func (h *LoggerCache) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h *LoggerCache) Fire(entry *logrus.Entry) error {
	msg, err := entry.String()
	h.addEntry(msg)
	return err
}

func (h *LoggerCache) addEntry(entry string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.entries) >= h.maxLines {
		h.entries = h.entries[1:]
	}
	h.entries = append(h.entries, entry)
	if h.out != nil {
		h.out.Write([]byte(entry))
	}
}

func (h *LoggerCache) GetEntries() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	entriesCopy := make([]string, len(h.entries))
	copy(entriesCopy, h.entries)
	return entriesCopy
}

func (h *LoggerCache) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.entries = make([]string, 0, h.maxLines)

}

func (h *LoggerCache) SetOutput(out io.Writer) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.out = out
}
