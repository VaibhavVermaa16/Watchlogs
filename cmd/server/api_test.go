package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestNewMux(t *testing.T) {
	mux := newMux()

	tests := []struct {
		route       string
		expectedHdl http.HandlerFunc
	}{
		{route: "/ingest", expectedHdl: ingest},
		{route: "/search", expectedHdl: search},
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", tt.route, nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		if w.Code == http.StatusNotFound {
			t.Errorf("expected route %s to be registered, but got 404 Not Found", tt.route)
		}
	}
}

func TestIngestEndpoint(t *testing.T) {
	t.Run("ingest valid log with POST request", func(t *testing.T) {
		tempfile, err := os.CreateTemp("", "test_logs_*.txt")
		if err != nil {
			t.Fatalf("failed to create temp file: %v", err)
		}
		defer os.Remove(tempfile.Name())

		originalFile := file
		OriginalLogs := logs
		defer func() {
			file = originalFile
			logs = OriginalLogs
		}()

		file = tempfile
		logs = []LogEntry{}

		payload := map[string]string{
			"level":   "info",
			"message": "test log message",
		}
		reqBody, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("failed to marshal request: %v", err)
		}

		request := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewReader(reqBody))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		ingest(response, request)

		if response.Code != http.StatusOK {
			t.Errorf("expected status 200 OK, got %d", response.Code)
		}
		if response.Body.String() != "ok" {
			t.Errorf("expected body 'ok', got '%s'", response.Body.String())
		}

		file.Seek(0, 0)
		logs = []LogEntry{}
		// Load logs from disk to verify persistence
		loadFromDisk()

		// Verify log was stored
		mu.Lock()
		if len(logs) != 1 {
			t.Errorf("expected 1 log entry, got %d", len(logs))
		} else if logs[0].Level != "info" || logs[0].Message != "test log message" {
			t.Errorf("log entry mismatch: got level=%s, message=%s", logs[0].Level, logs[0].Message)
		}
		mu.Unlock()
	})

	t.Run("ingest with invalid JSON", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewReader([]byte("invalid json")))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		ingest(response, request)

		if response.Code != http.StatusBadRequest {
			t.Errorf("expected status 400 Bad Request, got %d", response.Code)
		}
	})

	t.Run("ingest with non-POST request", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/ingest", nil)
		response := httptest.NewRecorder()

		ingest(response, request)

		if response.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405 Method Not Allowed, got %d", response.Code)
		}
	})
}

func TestSearchEndpoint(t *testing.T) {
	// Reset logs before tests
	mu.Lock()
	logs = []LogEntry{
		{Timestamp: time.Now(), Level: "info", Message: "test message 1"},
		{Timestamp: time.Now(), Level: "error", Message: "test message 2"},
	}
	mu.Unlock()

	t.Run("search returns logs as JSON", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/search", nil)
		response := httptest.NewRecorder()

		search(response, request)

		if response.Code != http.StatusOK {
			t.Errorf("expected status 200 OK, got %d", response.Code)
		}

		// Verify Content-Type header
		contentType := response.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
		}

		// Decode and verify response
		var returnedLogs []LogEntry
		if err := json.NewDecoder(response.Body).Decode(&returnedLogs); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(returnedLogs) != 2 {
			t.Errorf("expected 2 log entries, got %d", len(returnedLogs))
		}
	})

	t.Run("search with empty logs", func(t *testing.T) {
		mu.Lock()
		logs = []LogEntry{}
		mu.Unlock()

		request := httptest.NewRequest(http.MethodGet, "/search", nil)
		response := httptest.NewRecorder()

		search(response, request)

		if response.Code != http.StatusOK {
			t.Errorf("expected status 200 OK, got %d", response.Code)
		}

		var returnedLogs []LogEntry
		if err := json.NewDecoder(response.Body).Decode(&returnedLogs); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(returnedLogs) != 0 {
			t.Errorf("expected 0 log entries, got %d", len(returnedLogs))
		}
	})

	t.Run("search with non-GET request", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/search", nil)
		response := httptest.NewRecorder()

		search(response, request)

		if response.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405 Method Not Allowed, got %d", response.Code)
		}
	})
}

func TestLoadFromDisk(t *testing.T) {
	// Create a temporary file with sample log entries
	tempFile, err := os.CreateTemp("", "logs_*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	sampleLogs := []LogEntry{
		{Timestamp: time.Now(), Level: "info", Message: "log from disk 1"},
		{Timestamp: time.Now(), Level: "error", Message: "log from disk 2"},
	}

	for _, logEntry := range sampleLogs {
		data, _ := json.Marshal(logEntry)
		tempFile.Write(append(data, '\n'))
	}
	tempFile.Close()

	// Open the temp file for reading
	file, err = os.OpenFile(tempFile.Name(), os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("failed to open temp file: %v", err)
	}
	defer file.Close()

	// Load logs from disk
	loadFromDisk()

	// Verify logs were loaded correctly
	mu.Lock()
	if len(logs) != len(sampleLogs) {
		t.Errorf("expected %d log entries, got %d", len(sampleLogs), len(logs))
	} else {
		for i, logEntry := range sampleLogs {
			if logs[i].Level != logEntry.Level || logs[i].Message != logEntry.Message {
				t.Errorf("log entry mismatch at index %d: expected level=%s, message=%s; got level=%s, message=%s",
					i, logEntry.Level, logEntry.Message, logs[i].Level, logs[i].Message)
			}
		}
	}
	mu.Unlock()
}
