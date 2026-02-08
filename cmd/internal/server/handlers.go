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
	// Check if server is ready before processing the request
	if atomic.LoadInt64(&s.App.Metrics.Ready) == 0 {
		log.Printf("Received ingest request from %s but server is not ready\n", r.RemoteAddr)
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("server is not ready, try again later"))
		return
	}

	// Only allow POST requests for ingesting logs
	if r.Method != http.MethodPost {
		log.Printf("Received non-POST request on /ingest: %s\n", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	log.Printf("Received ingest request from %s\n", r.RemoteAddr)

	// Increment total ingested logs metric
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
		log.Printf("Log channel is full, rejecting request from %s\n", r.RemoteAddr)
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("log channel is full, try again later"))
	}
}

func (s *Server) Search(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadInt64(&s.App.Metrics.Ready) == 0 {
		log.Printf("Received search request from %s but server is not ready\n", r.RemoteAddr)
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("server is not ready, try again later"))
		return
	}
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
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("query cannot be empty"))
		return
	}

	s.App.Mu.Lock()
	defer s.App.Mu.Unlock()

	var results []app.LogEntry
	var ids []int
	sinceTime := helper.ParseSince(r.URL.Query().Get("since"))

	for seg := len(s.App.Segments) - 1; seg >= 0 && len(ids) < s.App.Cfg.MaxResults; seg-- {
		segment := s.App.Segments[seg]
		for i := range tokens {
			if len(ids) == 0 {
				ids = segment.Index[tokens[i]]
			} else {
				ids = helper.Intersect(ids, segment.Index[tokens[i]])
			}
		}

		for i := len(ids) - 1; i >= 0 && len(results) < s.App.Cfg.MaxResults; i-- {
			e := segment.Logs[ids[i]]
			if !sinceTime.IsZero() && e.Timestamp.Before(sinceTime) {
				continue
			}
			results = append(results, e)
		}

		// reset ids for next segment
		ids = nil
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
	var logCount = 0
	for _, seg := range s.App.Segments {
		logCount += len(seg.Logs)
	}
	var tokenCount = 0
	for _, ids := range s.App.CurrentSegment.Index {
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

func (s *Server) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Printf("Received non-GET request on /health: %s\n", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	log.Printf("Received health check request from %s\n", r.RemoteAddr)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (s *Server) Ready(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Printf("Received non-GET request on /ready: %s\n", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if atomic.LoadInt64(&s.App.Metrics.Ready) == 1 {
		log.Printf("Received ready check request from %s - READY\n", r.RemoteAddr)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
		return
	}

	log.Printf("Received ready check request from %s - NOT READY\n", r.RemoteAddr)
	w.WriteHeader(http.StatusServiceUnavailable)
	w.Write([]byte("not ready"))
}
