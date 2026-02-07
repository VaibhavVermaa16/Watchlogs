package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"watchlogs/cmd/helper"
	"watchlogs/cmd/internal/app"
)

func (s *Server) Ingest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("Received non-POST request on /ingest: %s\n", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	log.Printf("Received ingest request from %s\n", r.RemoteAddr)

	atomic.AddInt64(&s.App.Metrics.TotalIngested, 1)

	var req struct {
		Level   string `json:"level"`
		Message string `json:"message"`
	}

	if json.NewDecoder(r.Body).Decode(&req) != nil {
		log.Printf("Invalid request body from %s\n", r.RemoteAddr)
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	entry := app.LogEntry{
		Timestamp: time.Now(),
		Level:     req.Level,
		Message:   req.Message,
	}

	select {
	case s.App.LogCh <- entry:
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("ok"))

	default:
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("log channel is full, try again later"))
	}
}

func (s *Server) Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Printf("Received non-GET request on /search: %s\n", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	log.Printf("Received search request from %s with query: %s\n", r.RemoteAddr, r.URL.RawQuery)
	atomic.AddInt64(&s.App.Metrics.TotalSearched, 1)

	q := r.URL.Query().Get("q")

	tokens := helper.Tokenize(q)
	if len(tokens) == 0 {
		log.Printf("No tokens found in query: %s returning all logs\n", q)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.App.Logs)
		return
	}

	s.App.Mu.Lock()
	defer s.App.Mu.Unlock()
	log.Printf("Locked app state for search query: %s\n", q)
	var results []app.LogEntry
	// if q != "" {
	// 	maxResults := 5
	// 	ids := s.App.Index[q]
	// 	for i := len(ids) - 1; i >= 0 && len(results) < maxResults; i-- {
	// 		results = append(results, s.App.Logs[ids[i]])
	// 	}
	// } else {
	// 	results = s.App.Logs
	// }

	var ids []int

	for i := range tokens {
		if len(ids) == 0 {
			ids = s.App.Index[tokens[i]]
		} else {
			ids = helper.Intersect(ids, s.App.Index[tokens[i]])
		}
	}

	sinceTime := helper.ParseSince(r.URL.Query().Get("since"))

	for i := len(ids) - 1; i >= 0; i-- {
		e := s.App.Logs[ids[i]]
		if !sinceTime.IsZero() && e.Timestamp.Before(sinceTime) {
			continue
		}
		results = append(results, e)
	}

	// Return results as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (s *Server) Metrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Printf("Received non-GET request on /metrics: %s\n", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	uptime := time.Since(s.App.Metrics.StartTime).Seconds()
	s.App.Mu.Lock()
	var logCount = len(s.App.Logs)
	var tokenCount = 0
	for _, ids := range s.App.Index {
		tokenCount += len(ids)
	}
	s.App.Mu.Unlock()

	fmt.Fprintf(w,
		"uptime_sec %.0f\nlogs %d\ntokens %d\ningested %d\nsearched %d\n",
		uptime,
		logCount,
		tokenCount,
		atomic.LoadInt64(&s.App.Metrics.TotalIngested),
		atomic.LoadInt64(&s.App.Metrics.TotalSearched))
}
