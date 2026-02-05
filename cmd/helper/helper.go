package helper

import (
	"time"
	"watchlogs/cmd/internal/app"
)

const retentionPeriod = 24 * time.Hour

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
		return time.Now()
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
		cutoff := time.Now().Add(-retentionPeriod)

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
