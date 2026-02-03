package server

import (
	"slices"
	"encoding/json"
	"os"
	"testing"
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
		File: tempfile,
		Index: make(map[string][]int),
	}
	srv := New(a)

	if srv.App != a {
		t.Fatalf("expected server app to be set correctly")
	}

	payload := []app.LogEntry{
		{Level: "INFO", Message: "First log entry"},
		{Level: "ERROR", Message: "Second log entry"},
		{Level: "DEBUG", Message: "Third log entry"},
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
	srv.App.File = file

	srv.LoadFromDisk()

	// Check if logs are loaded correctly
	if len(srv.App.Logs) != len(payload) {
		t.Fatalf("expected %d log entries, got %d", len(payload), len(srv.App.Logs))
	}

	for i, logEntry := range payload {
		if srv.App.Logs[i].Level != logEntry.Level || srv.App.Logs[i].Message != logEntry.Message {
			t.Errorf("log entry mismatch at index %d: expected level=%s, message=%s; got level=%s, message=%s",
				i, logEntry.Level, logEntry.Message, srv.App.Logs[i].Level, srv.App.Logs[i].Message)
		}
	}
	
	// Check if index is built correctly
	for i, logEntry := range payload {
		tokens := helper.Tokenize(logEntry.Message)
		for _, token := range tokens {
			ids, exists := srv.App.Index[token]
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