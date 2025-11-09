package runner

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"syscall"
	"time"
)

// Runner starts/stops the child process and forwards stdio.
type Runner struct {
	mu  sync.Mutex
	cmd *exec.Cmd
}

// New returns a fresh Runner instance.
func New() *Runner { return &Runner{} }

// Start launches the given command using a shell.
func (r *Runner) Start(cmdStr string) error {
	if cmdStr == "" {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.cmd != nil && r.cmd.Process != nil {
		return nil
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", cmdStr)
	} else {
		cmd = exec.Command("sh", "-c", cmdStr)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}

	r.cmd = cmd

	// Wait in background and clear when done
	go func(c *exec.Cmd) {
		_ = c.Wait()
		r.mu.Lock()
		if r.cmd == c {
			r.cmd = nil
		}
		r.mu.Unlock()
	}(cmd)

	return nil
}

// Stop gracefully stops the process or kills it after a timeout.
func (r *Runner) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.cmd == nil || r.cmd.Process == nil {
		return
	}

	if runtime.GOOS == "windows" {
		_ = r.cmd.Process.Kill()
		r.cmd = nil
		return
	}

	_ = r.cmd.Process.Signal(syscall.SIGINT)

	done := make(chan struct{})
	go func() {
		_ = r.cmd.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		_ = r.cmd.Process.Kill()
	}

	r.cmd = nil
}
