package app

import (
	"os"
	"sync"
	"time"
)

type App struct {
	Mu             sync.Mutex
	LogCh          chan LogEntry
	Metrics        Metrics
	Cfg            Config
	CurrentSegment *Segment
	Segments       []*Segment
}

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"` // Write `json:"timestamp"` to specify JSON key because field name is capitalized in Go but should be lowercase in JSON
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

type Metrics struct {
	Ready         int64     `json:"ready"`
	TotalIngested int64     `json:"totalIngested"`
	TotalSearched int64     `json:"totalSearched"`
	StartTime     time.Time `json:"startTime"`
}

type Config struct {
	Retention   time.Duration
	MaxResults  int
	ChannelSize int
	DataPath    string
	MaxPerToken int
	MaxSegSize  int64
	HotSegments int
}

type Segment struct {
	Id    int
	File  *os.File
	Size  int64
	Logs  []LogEntry
	Index map[string][]int
}
