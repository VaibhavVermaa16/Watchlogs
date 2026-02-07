package server

import "net/http"

func (s *Server) Router() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/ingest", s.Ingest)
	mux.HandleFunc("/search", s.Search)
	mux.HandleFunc("/metrics", s.Metrics)
	mux.HandleFunc("/health", s.Health)
	mux.HandleFunc("/ready", s.Ready)
	return mux
}
