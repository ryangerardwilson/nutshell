package compiler

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

// Step counts completed milestones; it is not a time estimate.
type Progress struct {
	Stage   string
	Message string
	Step    int
}

func (o Options) progress(stage, message string, step int) {
	if o.OnProgress != nil {
		o.OnProgress(Progress{stage, message, step})
	}
}

func activitySummary(message string) (string, bool) {
	message = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			return ' '
		}
		return r
	}, message)
	words := strings.Fields(message)
	if len(words) < 3 || len(words) > 7 {
		return "", false
	}
	message = strings.Join(words, " ")
	return message, len([]rune(message)) <= 140
}

// Providers write progress.json independently of their chat/output format.
// A missing or partially written update must never break compilation.
func watchProgress(ctx context.Context, dir string, notify func(Progress)) func() {
	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()
		lastStep := 1
		var lastUpdate time.Time
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				path := filepath.Join(dir, "progress.json")
				info, err := os.Stat(path)
				if err != nil || info.ModTime().Equal(lastUpdate) {
					continue
				}
				var update struct {
					Version int    `json:"version"`
					Stage   string `json:"stage"`
					Message string `json:"message"`
				}
				if err := readJSON(path, &update); err != nil || update.Version != 1 {
					continue
				}
				step := map[string]int{"understanding": 1, "implementing": 2, "checking": 3, "ready": 4}[update.Stage]
				message, valid := activitySummary(update.Message)
				if step == 0 || !valid {
					continue
				}
				// An agent may fix implementation after testing. Show that work
				// without moving the milestone bar backward.
				if step < lastStep {
					step = lastStep
				}
				lastStep, lastUpdate = step, info.ModTime()
				if notify != nil {
					notify(Progress{update.Stage, message, step})
				}
			}
		}
	}()
	return func() { cancel(); <-done }
}
