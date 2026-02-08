package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
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

	entries, err := os.ReadDir(s.App.Cfg.DataPath)
	if err != nil {
		log.Printf("Failed to read data directory %s: %v\n", s.App.Cfg.DataPath, err)
	}

	var segIDs []int
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		var id int
		_, err := fmt.Sscanf(entry.Name(), "seg-%06d.log", &id)
		if err == nil {
			segIDs = append(segIDs, id)
		}
	}

	sort.Ints(segIDs)
	if len(segIDs) == 0 {
		seg, err := helper.OpenSegment(1, s.App.Cfg.DataPath)
		if err != nil {
			log.Fatal(err)
		}
		s.App.CurrentSegment = seg
		s.App.Segments = []*app.Segment{seg}
		return
	}

	hotCount := max(s.App.Cfg.HotSegments, 1)
	start := max(len(segIDs) - hotCount, 0)
	segIDs = segIDs[start:]

	var hotSegments []*app.Segment
	for i, id := range segIDs {
		seg, err := helper.OpenSegment(id, s.App.Cfg.DataPath)
		if err != nil {
			log.Printf("Failed to open segment %d: %v\n", id, err)
			continue
		}

		seg.Logs = nil
		seg.Index = make(map[string][]int)

		file, err := os.Open(filepath.Join(s.App.Cfg.DataPath, fmt.Sprintf("seg-%06d.log", id)))
		if err != nil {
			log.Printf("Failed to scan segment %d: %v\n", id, err)
			seg.File.Close()
			continue
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			var entry app.LogEntry
			if json.Unmarshal(scanner.Bytes(), &entry) != nil {
				continue
			}
			cutoff := time.Now().Add(-s.App.Cfg.Retention)
			if entry.Timestamp.After(cutoff) {
				logID := len(seg.Logs)
				seg.Logs = append(seg.Logs, entry)

				for _, token := range helper.Tokenize(entry.Message) {
					seg.Index[token] = append(seg.Index[token], logID)
				}
			}
		}
		file.Close()

		if i < len(segIDs)-1 {
			seg.File.Close()
		}
		hotSegments = append(hotSegments, seg)
	}

	if len(hotSegments) == 0 {
		seg, err := helper.OpenSegment(1, s.App.Cfg.DataPath)
		if err != nil {
			log.Fatal(err)
		}
		s.App.CurrentSegment = seg
		s.App.Segments = []*app.Segment{seg}
		return
	}

	s.App.Segments = hotSegments
	s.App.CurrentSegment = hotSegments[len(hotSegments)-1]
}
