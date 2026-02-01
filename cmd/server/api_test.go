package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	// Reset logs before tests to ensure isolation
	mu.Lock()
	logs = []LogEntry{}
	mu.Unlock()

	t.Run("ingest valid log with POST request", func(t *testing.T) {
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
