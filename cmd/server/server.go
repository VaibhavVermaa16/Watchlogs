package main

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"` // Write `json:"timestamp"` to specify JSON key because field name is capitalized in Go but should be lowercase in JSON
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

var (
	mu sync.Mutex // if we just want to make sure only one goroutine can access a variable at a time to avoid conflicts
	// It provides two methods: Lock and Unlock
	logs []LogEntry // Storing the recieved logs in memory for simplicity
)

func ingest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Defining a struct to parse incoming JSON log entry
	var req struct {
		Level   string `json:"level"`
		Message string `json:"message"`
	}

	// Decoding the JSON body into the struct
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// Creating a new log entry with the current timestamp
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     req.Level,
		Message:   req.Message,
	}

	// Storing the log entry in memory (thread-safe)
	mu.Lock()
	logs = append(logs, entry)
	mu.Unlock()
	w.Write([]byte("ok"))
}

func search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Retrieving logs (thread-safe)
	mu.Lock()
	defer mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	// Encoding logs to JSON and writing to response using json.NewEncoder iterates the JSON directly to the response writer without needing an intermediate buffer.
	if err := json.NewEncoder(w).Encode(logs); err != nil {
		http.Error(w, "failed to encode logs", http.StatusInternalServerError)
		return
	}
}
