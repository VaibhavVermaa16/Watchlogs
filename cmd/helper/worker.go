package helper

import (
	"encoding/json"
	"log"
	"watchlogs/cmd/internal/app"
)

func Writer(logCh <-chan app.LogEntry, a *app.App) {
	log.Println("Starting log writer goroutine...")
	for entry := range logCh {
		data, _ := json.Marshal(entry)
		a.Mu.Lock()

		id := len(a.Logs)
		log.Printf("Writing log entry with ID %d\n", id)
		a.Logs = append(a.Logs, entry)

		for _, token := range Tokenize(entry.Message) {
			ids := a.Index[token]
			if len(ids) >= a.Cfg.MaxPerToken {
				ids = ids[1:] // Remove oldest ID to maintain size
			}
			a.Index[token] = append(ids, id)
		}

		n, _ := a.CurrentSegment.File.Write(append(data, '\n'))
		a.CurrentSegment.Size += int64(n)

		// Check if we need to rotate the segment after writing
		if a.Cfg.MaxSegSize > 0 && a.CurrentSegment.Size >= a.Cfg.MaxSegSize {
			log.Printf("Current segment size %d exceeds max segment size %d, rotating segment...\n", a.CurrentSegment.Size, a.Cfg.MaxSegSize)

			a.CurrentSegment.File.Sync()
			a.CurrentSegment.File.Close()

			nextID := a.CurrentSegment.Id + 1
			newSeg, err := OpenSegment(nextID, a.Cfg.DataPath)
			if err != nil {
				log.Fatalf("Failed to open new segment: %v\n", err)
			}

			a.CurrentSegment = newSeg
			a.Segments = append(a.Segments, newSeg)
			log.Printf("Rotated to new segment with ID %d\n", nextID)
		}
		a.Mu.Unlock()
	}
}
