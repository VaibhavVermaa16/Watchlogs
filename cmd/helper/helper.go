package helper

import (
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

	path := "cmd/data/logs.txt"
	if v := os.Getenv("DATA_PATH"); v != "" {
		path = v
	}

	maxPerToken := 1000
	if v := os.Getenv("MAX_PER_TOKEN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			maxPerToken = n
		}
	}

	return app.Config{
		Retention:   ret,
		MaxResults:  maxRes,
		ChannelSize: chSize,
		MaxPerToken: maxPerToken,
		DataPath:    path,
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
		return time.Time{}
	}

	duration, err := time.ParseDuration(since)
	if err != nil {
		return time.Now()
	}

	return time.Now().Add(-duration)
}

func Cleanup(a *app.App) {
	ticker := time.NewTicker(10 * time.Minute)

	for range ticker.C {
		cutoff := time.Now().Add(-a.Cfg.Retention)

		// Perform cleanup logic here, e.g., remove old log entries from memory and disk

		var newLogs []app.LogEntry
		newIndex := make(map[string][]int)

		a.Mu.Lock()
		for _, log := range a.Logs {
			if log.Timestamp.After(cutoff) {
				id := len(newLogs)
				newLogs = append(newLogs, log)

				for _, token := range Tokenize(log.Message) {
					newIndex[token] = append(newIndex[token], id)
				}
			}
		}

		a.Index = newIndex
		a.Logs = newLogs
		a.Mu.Unlock()
	}
}
