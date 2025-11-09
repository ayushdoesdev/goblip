package watcher

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ayushdoesdev/goblip/internal/runner"
)

// Options holds configuration for the watcher.
type Options struct {
	Root       string
	Extensions string
	Interval   time.Duration
	Verbose    bool
	RunCommand string
}

// Watcher monitors files and restarts the runner on changes.
type Watcher struct {
	opts    Options
	runner  *runner.Runner
	stopCh  chan struct{}
	stopped chan struct{}
}

// New creates a new Watcher.
func New(opts Options, r *runner.Runner) *Watcher {
	if opts.Root == "" {
		opts.Root = "."
	}
	if opts.Interval <= 0 {
		opts.Interval = 500 * time.Millisecond
	}
	return &Watcher{
		opts:    opts,
		runner:  r,
		stopCh:  make(chan struct{}),
		stopped: make(chan struct{}),
	}
}

// Run starts the watcher loop.
func (w *Watcher) Run() error {
	extMap := parseExts(w.opts.Extensions)
	mtimes, err := scanFiles(w.opts.Root, extMap)
	if err != nil {
		return err
	}

	if err := w.runner.Start(w.opts.RunCommand); err != nil {
		return fmt.Errorf("initial start failed: %w", err)
	}

	ticker := time.NewTicker(w.opts.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			newTimes, _ := scanFiles(w.opts.Root, extMap)
			if changed(mtimes, newTimes) {
				if w.opts.Verbose {
					fmt.Println("[GoBlip] change detected — restarting")
				}
				mtimes = newTimes
				w.runner.Stop()
				time.Sleep(100 * time.Millisecond)
				if err := w.runner.Start(w.opts.RunCommand); err != nil {
					return fmt.Errorf("failed to restart command: %w", err)
				}
			}
		case <-w.stopCh:
			close(w.stopped)
			return nil
		}
	}
}

// Stop signals the watcher to stop.
func (w *Watcher) Stop() {
	select {
	case <-w.stopCh:
	default:
		close(w.stopCh)
		<-w.stopped
	}
}

// ---------------- Helpers ----------------

func parseExts(s string) map[string]struct{} {
	m := make(map[string]struct{})
	for _, e := range strings.Split(s, ",") {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		if !strings.HasPrefix(e, ".") {
			e = "." + e
		}
		m[e] = struct{}{}
	}
	return m
}

func scanFiles(root string, exts map[string]struct{}) (map[string]time.Time, error) {
	out := make(map[string]time.Time)
	if root == "" {
		root = "."
	}
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if base == ".git" || strings.HasPrefix(base, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if _, ok := exts[filepath.Ext(path)]; ok {
			out[path] = info.ModTime()
		}
		return nil
	})
	return out, nil
}

func changed(a, b map[string]time.Time) bool {
	if len(a) != len(b) {
		return true
	}
	for k, v := range b {
		if av, ok := a[k]; !ok || !av.Equal(v) {
			return true
		}
	}
	return false
}
