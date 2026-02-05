package app

import (
	"os"
	"sync"
	"time"
)

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"` // Write `json:"timestamp"` to specify JSON key because field name is capitalized in Go but should be lowercase in JSON
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

type App struct {
	File  *os.File
	Logs  []LogEntry
	Index map[string][]int
	Mu    sync.Mutex
	LogCh chan LogEntry
}
