package runner

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Runner manages a child process started from a shell command string.
type Runner struct {
	CmdStr  string
	Verbose bool

	mu  sync.Mutex
	cmd *exec.Cmd
}

// extractPort attempts to extract the port number from the command string
func (r *Runner) extractPort() int {
	// Common port patterns
	patterns := []string{":8080", "PORT=", "-p ", "--port="}

	for _, pattern := range patterns {
		if idx := strings.Index(r.CmdStr, pattern); idx != -1 {
			// Extract the port number
			portStr := r.CmdStr[idx+len(pattern):]
			portStr = strings.Split(portStr, " ")[0]
			if port, err := strconv.Atoi(strings.TrimPrefix(portStr, ":")); err == nil {
				return port
			}
		}
	}
	return 8080 // default port for many web servers
}

// waitForPortRelease waits until the specified port is available
func (r *Runner) waitForPortRelease(port int) {
	start := time.Now()
	for time.Since(start) < 5*time.Second {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf(":%d", port), 100*time.Millisecond)
		if err != nil {
			// Port is available
			return
		}
		if conn != nil {
			conn.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}
	if r.Verbose {
		fmt.Printf("[gowatch] warning: port %d might still be in use\n", port)
	}
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

	// Wait for port to be released before starting
	port := r.extractPort()
	r.waitForPortRelease(port)

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", r.CmdStr)
	} else {
		cmd = exec.Command("sh", "-c", r.CmdStr)
	}

	// Set process group for better process management
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

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
		// Send SIGTERM to process group
		pgid, err := syscall.Getpgid(r.cmd.Process.Pid)
		if err == nil {
			_ = syscall.Kill(-pgid, syscall.SIGTERM)
		}

		// Wait for graceful shutdown
		done := make(chan struct{})
		go func() {
			r.cmd.Wait()
			close(done)
		}()

		// Allow more time for graceful shutdown of servers
		select {
		case <-done:
			if r.Verbose {
				fmt.Println("[gowatch] process exited gracefully")
			}
		case <-time.After(3 * time.Second):
			if r.Verbose {
				fmt.Println("[gowatch] forcing process termination")
			}
			if pgid != 0 {
				_ = syscall.Kill(-pgid, syscall.SIGKILL)
			}
			_ = r.cmd.Process.Kill()
		}
	}

	r.cmd = nil

	// Add a small delay after stopping to ensure cleanup
	time.Sleep(500 * time.Millisecond)
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
