// gowatch.go
//
// Simple nodemon-like watcher for Go projects using only the standard library.
// Defaults to running `go run .` if no command is provided.
//
// Usage examples:
//   go run gowatch.go                 # runs `go run .` and restarts on changes
//   go run gowatch.go -- go run main.go
//   go run gowatch.go -- -cmd "mybinary -flag"
//   go build -o gowatch && ./gowatch -- go run ./cmd/myapp
//
// Notes:
// - Uses polling (interval default 500ms) so it works cross-platform without fsnotify.
// - Watches .go, .mod, .sum, .tpl, .html, .css, .js files by default. Change `extensions` if desired.

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ayushdoesdev/goblip/internal/runner"
	"github.com/ayushdoesdev/goblip/internal/watcher"
)

var (
	intervalFlag = flag.Duration("interval", 500*time.Millisecond, "poll interval for file changes")
	extsFlag     = flag.String("ext", ".go,.mod,.sum,.tpl,.html,.css,.js", "comma-separated list of file extensions to watch")
	ignoreVcs    = flag.Bool("ignore-vcs", true, "ignore .git, .hg, .svn directories")
	verbose      = flag.Bool("v", false, "verbose output")
)

func main() {
	flag.Parse()

	cmdArgs := flag.Args()
	var runCmd string
	if len(cmdArgs) == 0 {
		runCmd = "go run ."
	} else {
		runCmd = strings.Join(cmdArgs, " ")
	}

	exts := watcher.ParseExts(*extsFlag)

	if *verbose {
		fmt.Printf("Watching extensions: %v\n", exts)
		fmt.Printf("Command to run: %s\n", runCmd)
		fmt.Printf("Poll interval: %v\n", *intervalFlag)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w := watcher.New(*intervalFlag, exts, *ignoreVcs, *verbose)
	restartCh, err := w.Start(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "watcher start error: %v\n", err)
		os.Exit(1)
	}

	r := runner.New(runCmd, *verbose)
	if err := r.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start command: %v\n", err)
		os.Exit(1)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case <-restartCh:
			if *verbose {
				fmt.Println("[gowatch] restarting child")
			}
			r.Stop()
			time.Sleep(100 * time.Millisecond)
			if err := r.Start(); err != nil {
				fmt.Fprintf(os.Stderr, "failed to restart command: %v\n", err)
			}
		case s := <-sigs:
			if *verbose {
				fmt.Printf("[gowatch] received signal: %v\n", s)
			}
			r.Signal(s)
			return
		}
	}
}
