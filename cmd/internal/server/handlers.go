package server

import (
	"encoding/json"
	"net/http"
	"time"

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

	s.App.Mu.Lock()
	defer s.App.Mu.Unlock()

	var results []app.LogEntry
	if q != "" {
		for _, id := range s.App.Index[q] {
			results = append(results, s.App.Logs[id])
		}
	} else {
		results = s.App.Logs
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
