package main

import (
	"net/http"
)

func ingest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Write([]byte("ingested"))
}

func search(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("search endpoint"))
}
