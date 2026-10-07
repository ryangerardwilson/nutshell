package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/ryangerardwilson/nutshell/internal/compiler"
)

type progressDisplay struct {
	mu       sync.Mutex
	w        io.Writer
	terminal bool
	current  compiler.Progress
	started  time.Time
	stop     chan struct{}
	done     chan struct{}
	frame    int
	updated  time.Time
}

func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok || os.Getenv("TERM") == "dumb" {
		return false
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func newProgressDisplay(w io.Writer, terminal bool) *progressDisplay {
	p := &progressDisplay{w: w, terminal: terminal, started: time.Now(), updated: time.Now(), stop: make(chan struct{}), done: make(chan struct{}), current: compiler.Progress{Stage: "preparing", Message: "Preparing your build workspace"}}
	go func() {
		defer close(p.done)
		ticker := time.NewTicker(120 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-p.stop:
				return
			case <-ticker.C:
				if terminal {
					p.mu.Lock()
					p.render(false)
					p.mu.Unlock()
				}
			}
		}
	}()
	return p
}

func cleanMessage(s string) string {
	s = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			return ' '
		}
		return r
	}, s)
	return strings.Join(strings.Fields(s), " ")
}

func (p *progressDisplay) update(update compiler.Progress) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if update.Step < p.current.Step {
		return
	}
	changed := update.Stage != p.current.Stage || cleanMessage(update.Message) != cleanMessage(p.current.Message)
	p.current = update
	p.updated = time.Now()
	if p.terminal {
		p.render(false)
	} else if changed {
		fmt.Fprintf(p.w, "%s\n", cleanMessage(update.Message))
	}
}

func (p *progressDisplay) render(final bool) {
	filled := p.current.Step * 16 / 8
	if filled < 0 {
		filled = 0
	}
	if filled > 16 {
		filled = 16
	}
	marker := string("|/-\\"[p.frame%4])
	p.frame++
	if final {
		marker = " "
	}
	fmt.Fprintf(p.w, "\r\x1b[K[%s%s] %s %s %s", strings.Repeat("=", filled), strings.Repeat(" ", 16-filled), marker, p.activity(time.Now(), final), time.Since(p.started).Truncate(time.Second))
	if final {
		fmt.Fprintln(p.w)
	}
}

// Only the AI generation phase needs a stale-report fallback. Build and test
// commands have known activity and may legitimately run for several minutes.
func (p *progressDisplay) activity(now time.Time, final bool) string {
	if !final && p.current.Step >= 1 && p.current.Step <= 4 && now.Sub(p.updated) >= 30*time.Second {
		return "Waiting for AI progress update"
	}
	return cleanMessage(p.current.Message)
}

func (p *progressDisplay) finish(err error) {
	close(p.stop)
	<-p.done
	p.mu.Lock()
	defer p.mu.Unlock()
	if err == nil {
		p.current = compiler.Progress{Stage: "done", Message: "Done", Step: 8}
	} else {
		p.current.Message = "Build stopped"
	}
	if p.terminal {
		p.render(true)
	} else {
		fmt.Fprintln(p.w, p.current.Message)
	}
}
