package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-stop // Wait for shutdown signal, program pause here until signal is received
		log.Println("shutting down server...")
		close(a.LogCh) // Close the log channel to stop the writer goroutine
		file.Sync()    // Ensure all data is flushed to disk
		file.Close()
		os.Exit(0)
	}()

	// Log rotation to remove unwanted old logs
	go helper.Cleanup(a)

	log.Println("server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", srv.Router()))
}
