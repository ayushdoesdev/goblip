package runner

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

// Runner manages a child process started from a shell command string.
type Runner struct {
	CmdStr  string
	Verbose bool

	mu  sync.Mutex
	cmd *exec.Cmd
}

func New(cmdStr string, verbose bool) *Runner {
	return &Runner{CmdStr: cmdStr, Verbose: verbose}
}

// Start launches the configured command (uses a shell) and returns any start error.
func (r *Runner) Start() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cmd != nil && r.cmd.Process != nil {
		return fmt.Errorf("process already running")
	}
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", r.CmdStr)
	} else {
		cmd = exec.Command("sh", "-c", r.CmdStr)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if r.Verbose {
		fmt.Printf("[gowatch] starting: %s\n", r.CmdStr)
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	r.cmd = cmd
	go func(cmd *exec.Cmd) {
		err := cmd.Wait()
		if err != nil && r.Verbose {
			fmt.Fprintf(os.Stderr, "[gowatch] child exited with error: %v\n", err)
		} else if r.Verbose {
			fmt.Printf("[gowatch] child exited\n")
		}
		r.mu.Lock()
		if r.cmd == cmd {
			r.cmd = nil
		}
		r.mu.Unlock()
	}(cmd)
	return nil
}

// Stop attempts a graceful shutdown then kills the process if it doesn't exit quickly.
func (r *Runner) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cmd == nil || r.cmd.Process == nil {
		return
	}
	if r.Verbose {
		fmt.Printf("[gowatch] stopping pid %d\n", r.cmd.Process.Pid)
	}
	if runtime.GOOS == "windows" {
		_ = r.cmd.Process.Kill()
	} else {
		_ = r.cmd.Process.Signal(os.Interrupt)
		done := make(chan struct{})
		go func() {
			r.cmd.Wait()
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			_ = r.cmd.Process.Kill()
		}
	}
	r.cmd = nil
}

// Signal forwards an OS signal to the child and then ensures it is killed shortly after.
func (r *Runner) Signal(sig os.Signal) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cmd == nil || r.cmd.Process == nil {
		return
	}
	_ = r.cmd.Process.Signal(sig)
	time.Sleep(500 * time.Millisecond)
	_ = r.cmd.Process.Kill()
}
