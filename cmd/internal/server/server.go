package server

import (
	"bufio"
	"encoding/json"
	"log"
	"time"

	"watchlogs/cmd/helper"
	"watchlogs/cmd/internal/app"
)

type Server struct {
	App *app.App
}

func New(a *app.App) *Server {
	return &Server{App: a}
}

func (s *Server) LoadFromDisk() {
	log.Println("Loading logs from disk...")
	scanner := bufio.NewScanner(s.App.File)
	s.App.Logs = nil

	for scanner.Scan() {
		var entry app.LogEntry
		if json.Unmarshal(scanner.Bytes(), &entry) != nil {
			continue
		}
		cutoff := time.Now().Add(-s.App.Cfg.Retention)
		if entry.Timestamp.After(cutoff) {
			id := len(s.App.Logs)
			s.App.Logs = append(s.App.Logs, entry)

			for _, token := range helper.Tokenize(entry.Message) {
				s.App.Index[token] = append(s.App.Index[token], id)
			}
		}
	}
}
