package main

import (
	"log"
	"net/http"
	"os"
)

func newMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/ingest", ingest)
	mux.HandleFunc("/search", search)

	return mux
}

func main() {
	mux := newMux()
	var err error
	file, err = os.OpenFile("cmd/data/logs.txt", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}

	loadFromDisk()
	log.Println("server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
