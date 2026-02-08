package helper

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	"watchlogs/cmd/internal/app"
)

func LoadConfig() app.Config {
	ret := 24 * time.Hour
	if v := os.Getenv("RETENTION"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			ret = d
		}
	}

	maxRes := 100
	if v := os.Getenv("MAX_RESULTS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			maxRes = n
		}
	}

	chSize := 1000
	if v := os.Getenv("CHANNEL_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			chSize = n
		}
	}

	path := "cmd/data/"
	if v := os.Getenv("DATA_PATH"); v != "" {
		path = v
	}

	maxPerToken := 1000
	if v := os.Getenv("MAX_PER_TOKEN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			maxPerToken = n
		}
	}

	maxSegSize := int64(10 * 1024 * 1024) // Default 10 MB
	if v := os.Getenv("MAX_SEG_SIZE"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			maxSegSize = n
		}
	}

	hotSegments := 2
	if v := os.Getenv("HOT_SEGMENTS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			hotSegments = n
		}
	}

	return app.Config{
		Retention:   ret,
		MaxResults:  maxRes,
		ChannelSize: chSize,
		MaxPerToken: maxPerToken,
		MaxSegSize:  maxSegSize,
		DataPath:    path,
		HotSegments: hotSegments,
	}
}

func Tokenize(input string) []string {
	var tokens []string
	current := ""

	for _, char := range input {
		if char >= 'A' && char <= 'Z' {
			char += 'a' - 'A' // Convert to lowercase
		}
		if char >= 'a' && char <= 'z' {
			current += string(char)
		} else {
			if current != "" {
				tokens = append(tokens, current)
				current = ""
			}
		}
	}
	if current != "" {
		tokens = append(tokens, current)
	}
	return tokens
}

func Intersect(a, b []int) []int {
	var i = 0
	var j = 0
	var result []int

	for i < len(a) && j < len(b) {
		if a[i] == b[j] {
			result = append(result, a[i])
			i++
			j++
		} else if a[i] < b[j] {
			i++
		} else {
			j++
		}
	}
	return result
}

func ParseSince(since string) time.Time {
	if since == "" {
		log.Println("No 'since' parameter provided, returning zero time")
		return time.Time{}
	}

	duration, err := time.ParseDuration(since)
	if err != nil {
		log.Printf("Invalid 'since' parameter: %s, error: %v\n", since, err)
		return time.Now()
	}

	return time.Now().Add(-duration)
}

func Cleanup(a *app.App) {
	ticker := time.NewTicker(1 * time.Hour)

	for range ticker.C {
		log.Println("Starting cleanup goroutine...")
		cutoff := time.Now().Add(-a.Cfg.Retention)

		a.Mu.Lock()
		var keptSegments []*app.Segment
		for _, segment := range a.Segments {
			shouldDelete := false
			if len(segment.Logs) > 0 {
				shouldDelete = segment.Logs[len(segment.Logs)-1].Timestamp.Before(cutoff)
			} else if segment.File != nil {
				if info, err := segment.File.Stat(); err == nil {
					shouldDelete = info.ModTime().Before(cutoff)
				}
			}

			if shouldDelete {
				if segment.File != nil {
					segment.File.Sync()
					segment.File.Close()
					_ = os.Remove(segment.File.Name())
				}

				if a.CurrentSegment == segment {
					nextID := segment.Id + 1
					newSeg, err := OpenSegment(nextID, a.Cfg.DataPath)
					if err != nil {
						log.Printf("Failed to open new segment after cleanup: %v\n", err)
					} else {
						a.CurrentSegment = newSeg
						keptSegments = append(keptSegments, newSeg)
					}
				}
				continue
			}
			keptSegments = append(keptSegments, segment)
		}

		if len(keptSegments) == 0 {
			newSeg, err := OpenSegment(1, a.Cfg.DataPath)
			if err != nil {
				log.Printf("Failed to open fallback segment after cleanup: %v\n", err)
			} else {
				a.CurrentSegment = newSeg
				keptSegments = append(keptSegments, newSeg)
			}
		}

		a.Segments = keptSegments
		a.Mu.Unlock()
		log.Println("Cleanup completed.")
	}
	log.Println("Cleanup goroutine stopped.")
}

func OpenSegment(id int, path string) (*app.Segment, error) {
	name := fmt.Sprintf("%s/seg-%06d.log", path, id)
	f, err := os.OpenFile(name, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	info, _ := f.Stat()
	return &app.Segment{
		Id:    id,
		File:  f,
		Size:  info.Size(),
		Index: make(map[string][]int),
	}, nil
}
