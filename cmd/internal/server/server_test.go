package server

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
	"time"
	"watchlogs/cmd/helper"
	"watchlogs/cmd/internal/app"
)

func TestServer(t *testing.T) {
	tempfile, err := os.CreateTemp("", "logs.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempfile.Name())
	defer tempfile.Close()
	a := &app.App{
		Cfg: helper.LoadConfig(),
		CurrentSegment: &app.Segment{
			Index: make(map[string][]int),
		},
	}
	srv := New(a)

	if srv.App != a {
		t.Fatalf("expected server app to be set correctly")
	}

	now := time.Now()
	payload := []app.LogEntry{
		{Timestamp: now, Level: "INFO", Message: "First log entry"},
		{Timestamp: now, Level: "ERROR", Message: "Second log entry"},
		{Timestamp: now, Level: "DEBUG", Message: "Third log entry"},
	}

	for _, logEntry := range payload {
		data, _ := json.Marshal(logEntry)
		tempfile.Write(append(data, '\n'))
	}
	tempfile.Close()
	tempfile.Seek(0, 0)

	file, err := os.Open(tempfile.Name())
	if err != nil {
		t.Fatalf("failed to open temp file: %v", err)
	}
	defer file.Close()
	srv.App.CurrentSegment.File = file

	srv.LoadFromDisk()

	// Check if logs are loaded correctly
	if len(srv.App.CurrentSegment.Logs) != len(payload) {
		t.Fatalf("expected %d log entries, got %d", len(payload), len(srv.App.CurrentSegment.Logs))
	}

	for i, logEntry := range payload {
		if srv.App.CurrentSegment.Logs[i].Level != logEntry.Level || srv.App.CurrentSegment.Logs[i].Message != logEntry.Message {
			t.Errorf("log entry mismatch at index %d: expected level=%s, message=%s; got level=%s, message=%s",
				i, logEntry.Level, logEntry.Message, srv.App.CurrentSegment.Logs[i].Level, srv.App.CurrentSegment.Logs[i].Message)
		}
	}

	// Check if index is built correctly
	for i, logEntry := range payload {
		tokens := helper.Tokenize(logEntry.Message)
		for _, token := range tokens {
			ids, exists := srv.App.CurrentSegment.Index[token]
			if !exists {
				t.Errorf("expected token %s to exist in index", token)
				continue
			}
			found := slices.Contains(ids, i)
			if !found {
				t.Errorf("expected log ID %d to be indexed under token %s", i, token)
			}
		}
	}
}
