package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"watchlogs/cmd/helper"
	"watchlogs/cmd/internal/app"
)

func TestIngest(t *testing.T) {
	t.Run("ingest valid payload", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "test_logs_*.txt")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()
		a := &app.App{
			File:  tempFile,
			Index: make(map[string][]int),
		}

		srv := New(a)

		payload := map[string]string{
			"level":   "INFO",
			"message": "This is a test log message",
		}

		body, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}

		req := httptest.NewRequest(http.MethodPost, "/ingest", io.NopCloser(bytes.NewReader(body)))
		res := httptest.NewRecorder()

		srv.Ingest(res, req)

		if res.Code != http.StatusOK {
			t.Fatalf("expected status 200 OK, got %d", res.Code)
		}

		// rewind file
		tempFile.Seek(0, 0)

		if len(a.Logs) != 1 {
			t.Fatalf("expected 1 log entry, got %d", len(a.Logs))
		}

		if a.Logs[0].Level != "INFO" || a.Logs[0].Message != "This is a test log message" {
			t.Fatalf("log entry does not match ingested data")
		}
	})
	t.Run("invalid json payload", func(t *testing.T) {
		a := &app.App{
			Index: make(map[string][]int),
		}
		srv := New(a)

		req := httptest.NewRequest(http.MethodPost, "/ingest", nil)
		res := httptest.NewRecorder()

		srv.Ingest(res, req)
		if res.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400 Bad Request, got %d", res.Code)
		}
	})
	t.Run("ingest with non-POST request", func(t *testing.T) {
		a := &app.App{
			Index: make(map[string][]int),
		}
		srv := New(a)
		request := httptest.NewRequest(http.MethodGet, "/ingest", nil)
		response := httptest.NewRecorder()

		srv.Ingest(response, request)

		if response.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405 Method Not Allowed, got %d", response.Code)
		}
	})
}

func TestSearch(t *testing.T) {
	a := &app.App{
		Index: make(map[string][]int),
	}
	srv := New(a)

	// Preload some logs
	a.Logs = []app.LogEntry{
		{Level: "INFO", Message: "first test log"},
		{Level: "ERROR", Message: "second test log"},
		{Level: "INFO", Message: "third log entry"},
	}
	for i, log := range a.Logs {
		for _, token := range helper.Tokenize(log.Message) {
			a.Index[token] = append(a.Index[token], i)
		}
	}

	t.Run("search returns logs as JSON", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/search?q=test", nil)
		response := httptest.NewRecorder()

		srv.Search(response, request)

		if response.Code != http.StatusOK {
			t.Errorf("expected status 200 OK, got %d", response.Code)
		}

		// Verify Content-Type header
		contentType := response.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
		}

		// Decode and verify response
		var returnedLogs []app.LogEntry
		if err := json.NewDecoder(response.Body).Decode(&returnedLogs); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		// Logs whose message contains "test" should be returned
		expectedCount := 2 // "first test log" and "second test log"
		if len(returnedLogs) != expectedCount {
			t.Errorf("expected %d log entries, got %d", expectedCount, len(returnedLogs))
		}

		// Verify the messages contain "test"
		for _, log := range returnedLogs {
			if log.Message != "first test log" && log.Message != "second test log" {
				t.Errorf("unexpected log message: %s", log.Message)
			}
		}
	})

	t.Run("search with non-GET request", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/search?q=INFO", nil)
		response := httptest.NewRecorder()

		srv.Search(response, request)

		if response.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405 Method Not Allowed, got %d", response.Code)
		}
	})

	t.Run("search with no matching logs", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/search?q=database", nil)
		response := httptest.NewRecorder()

		srv.Search(response, request)

		if response.Code != http.StatusOK {
			t.Errorf("expected status 200 OK, got %d", response.Code)
		}

		var returnedLogs []app.LogEntry
		if err := json.NewDecoder(response.Body).Decode(&returnedLogs); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(returnedLogs) != 0 {
			t.Errorf("expected 0 log entries, got %d", len(returnedLogs))
		}
	})
}
