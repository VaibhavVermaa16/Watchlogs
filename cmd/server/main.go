package main

import (
	"log"
	"net/http"
	"os"

	"watchlogs/cmd/helper"
	"watchlogs/cmd/internal/app"
	"watchlogs/cmd/internal/server"
)

func main() {
	file, err := os.OpenFile("cmd/data/logs.txt", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}

	a := &app.App{
		File:  file,
		Index: make(map[string][]int),
		LogCh: make(chan app.LogEntry, 1000),
	}

	srv := server.New(a)
	srv.LoadFromDisk()
	go helper.Writer(a.LogCh, a)

	log.Println("server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", srv.Router()))
}
