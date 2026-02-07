package helper

import (
	"encoding/json"
	"watchlogs/cmd/internal/app"
)

func Writer(logCh <-chan app.LogEntry, a *app.App) {
	// Example worker function that could perform background tasks
	for entry := range logCh {
		data, _ := json.Marshal(entry)
		a.Mu.Lock()

		id := len(a.Logs)
		a.Logs = append(a.Logs, entry)

		for _, token := range Tokenize(entry.Message) {
			ids := a.Index[token]
			if len(ids) >= a.Cfg.MaxPerToken {
				ids = ids[1:] // Remove oldest ID to maintain size
			}
			a.Index[token] = append(ids, id)
		}

		a.File.Write(append(data, '\n'))
		a.Mu.Unlock()
	}
}
