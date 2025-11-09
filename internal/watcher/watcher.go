package watcher

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Watcher polls the filesystem and emits restart notifications when files change.
type Watcher struct {
	Interval  time.Duration
	Exts      map[string]struct{}
	IgnoreVcs bool
	Verbose   bool
}

func New(interval time.Duration, exts map[string]struct{}, ignoreVcs, verbose bool) *Watcher {
	return &Watcher{
		Interval:  interval,
		Exts:      exts,
		IgnoreVcs: ignoreVcs,
		Verbose:   verbose,
	}
}

// ParseExts turns a comma-separated list into a map for quick checks
func ParseExts(s string) map[string]struct{} {
	out := make(map[string]struct{})
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !strings.HasPrefix(p, ".") {
			p = "." + p
		}
		out[p] = struct{}{}
	}
	return out
}

// Start begins watching and returns a channel that receives when a restart should occur.
// The channel is closed when ctx is cancelled.
func (w *Watcher) Start(ctx context.Context) (<-chan struct{}, error) {
	restartCh := make(chan struct{}, 1)

	mtimes, err := scanFiles(".", w.Exts, w.IgnoreVcs)
	if err != nil {
		return nil, fmt.Errorf("initial scan error: %w", err)
	}

	go func() {
		ticker := time.NewTicker(w.Interval)
		defer ticker.Stop()
		defer close(restartCh)
		for {
			select {
			case <-ticker.C:
				new, err := scanFiles(".", w.Exts, w.IgnoreVcs)
				if err != nil {
					if w.Verbose {
						fmt.Fprintf(os.Stderr, "scan error: %v\n", err)
					}
					continue
				}
				if changed(mtimes, new) {
					if w.Verbose {
						fmt.Println("change detected, signaling restart")
					}
					mtimes = new
					select {
					case restartCh <- struct{}{}:
					default:
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return restartCh, nil
}

// scanFiles walks directory and records mod times for files with matching extensions
func scanFiles(root string, exts map[string]struct{}, ignoreVcs bool) (map[string]time.Time, error) {
	out := make(map[string]time.Time)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// skip files we can't stat
			return nil
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if ignoreVcs && (base == ".git" || base == ".hg" || base == ".svn") {
				return filepath.SkipDir
			}
			if strings.HasPrefix(base, ".") && base != "." {
				return filepath.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(path)
		if _, ok := exts[ext]; ok {
			out[path] = info.ModTime()
		}
		return nil
	})
	return out, err
}

// changed compares two mtimes maps; returns true if any file was added, removed, or modtime changed.
func changed(old, now map[string]time.Time) bool {
	if len(old) != len(now) {
		return true
	}
	for k, v := range now {
		if ov, ok := old[k]; !ok {
			return true
		} else if !v.Equal(ov) {
			return true
		}
	}
	return false
}
