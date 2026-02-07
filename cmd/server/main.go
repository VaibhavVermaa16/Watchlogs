package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"watchlogs/cmd/helper"
	"watchlogs/cmd/internal/app"
	"watchlogs/cmd/internal/server"

	"github.com/joho/godotenv"
)

func main() {
	log.Println("Server started...")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	cfg := helper.LoadConfig()
	log.Printf("Config: %+v\n", cfg)
	file, err := os.OpenFile(cfg.DataPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}

	a := &app.App{
		File:  file,
		Index: make(map[string][]int),
		LogCh: make(chan app.LogEntry, cfg.ChannelSize),
		Cfg:   cfg,
	}
	a.Metrics.StartTime = time.Now()
	atomic.StoreInt64(&a.Metrics.Ready, 1)

	srv := server.New(a)
	srv.LoadFromDisk()
	go helper.Writer(a.LogCh, a)

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-stop // Wait for shutdown signal, program pause here until signal is received
		log.Println("shutting down server...")
		atomic.StoreInt64(&a.Metrics.Ready, 0)
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
