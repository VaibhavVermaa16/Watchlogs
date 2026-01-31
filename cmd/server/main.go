package main

import (
	"log"
	"net/http"
)

func newMux () *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/ingest", ingest)
	mux.HandleFunc("/search", search)

	return mux
}

func main() {
	mux := newMux()
	log.Println("server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}