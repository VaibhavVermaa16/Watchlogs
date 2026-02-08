package helper

import (
	"os"
	"testing"
	"time"
	"watchlogs/cmd/internal/app"
)

func TestWriter(t *testing.T) {
	cfg := LoadConfig()
	tempFile, err := os.CreateTemp("", "seg-*.log")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())
	defer tempFile.Close()

	a := &app.App{
		Cfg:   cfg,
		LogCh: make(chan app.LogEntry, cfg.ChannelSize),
		CurrentSegment: &app.Segment{
			Id:    1,
			File:  tempFile,
			Index: make(map[string][]int),
		},
	}

	go Writer(a.LogCh, a)
	entry := app.LogEntry{
		Timestamp: time.Now(),
		Level:     "info",
		Message:   "test log entry",
	}

	a.LogCh <- entry
	time.Sleep(200 * time.Millisecond) // Wait for the writer to process

	if len(a.CurrentSegment.Logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(a.CurrentSegment.Logs))
	}

	if a.CurrentSegment.Logs[0].Message != "test log entry" {
		t.Errorf("Expected message 'test log entry', got '%s'", a.CurrentSegment.Logs[0].Message)
	}
}
