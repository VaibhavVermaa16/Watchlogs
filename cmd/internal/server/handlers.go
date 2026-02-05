package server

import (
	"encoding/json"
	"net/http"
	"time"

	"watchlogs/cmd/helper"
	"watchlogs/cmd/internal/app"
)

func (s *Server) Ingest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Level   string `json:"level"`
		Message string `json:"message"`
	}

	if json.NewDecoder(r.Body).Decode(&req) != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	entry := app.LogEntry{
		Timestamp: time.Now(),
		Level:     req.Level,
		Message:   req.Message,
	}

	// data, _ := json.Marshal(entry)

	// s.App.Mu.Lock()
	// id := len(s.App.Logs)
	// s.App.Logs = append(s.App.Logs, entry)

	// for _, token := range helper.Tokenize(entry.Message) {
	// 	s.App.Index[token] = append(s.App.Index[token], id)
	// }

	// s.App.File.Write(append(data, '\n'))
	// s.App.File.Sync()
	// s.App.Mu.Unlock()
	s.App.LogCh <- entry

	w.Write([]byte("ok"))
}

func (s *Server) Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	q := r.URL.Query().Get("q")

	tokens := helper.Tokenize(q)
	if len(tokens) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.App.Logs)
		return
	}

	s.App.Mu.Lock()
	defer s.App.Mu.Unlock()
	
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

	for i := len(ids) - 1; i >= 0; i-- {
		results = append(results, s.App.Logs[ids[i]])
	}

	// Return results as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
