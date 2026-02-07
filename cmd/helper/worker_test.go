package helper

import (
	"testing"
	"time"
	"watchlogs/cmd/internal/app"
)

func TestWriter(t *testing.T) {
	cfg := LoadConfig()
	a := &app.App{
		Cfg:   cfg,
		Logs:  []app.LogEntry{},
		Index: make(map[string][]int),
		LogCh: make(chan app.LogEntry, cfg.ChannelSize),
	}

	go Writer(a.LogCh, a)
	entry := app.LogEntry{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "test log entry",
	}

	a.LogCh <- entry
	time.Sleep(200 * time.Millisecond) // Wait for the writer to process

	if len(a.Logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(a.Logs))
	}

	if a.Logs[0].Message != "test log entry" {
		t.Errorf("Expected message 'test log entry', got '%s'", a.Logs[0].Message)
	}
}
