package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ayushdoesdev/goblip/internal/runner"
	"github.com/ayushdoesdev/goblip/internal/watcher"
)

func main() {
	interval := flag.Duration("interval", 500*time.Millisecond, "poll interval for file changes")
	exts := flag.String("ext", ".go,.mod,.sum,.html,.css,.js", "comma-separated file extensions to watch")
	verbose := flag.Bool("v", false, "enable verbose output")
	flag.Parse()

	cmdArgs := flag.Args()
	runCmd := ""
	if len(cmdArgs) > 0 {
		for i, a := range cmdArgs {
			if i > 0 {
				runCmd += " "
			}
			runCmd += a
		}
	}

	opts := watcher.Options{
		Root:       ".",
		Extensions: *exts,
		Interval:   *interval,
		Verbose:    *verbose,
		RunCommand: runCmd,
	}

	r := runner.New()
	w := watcher.New(opts, r)

	// Handle OS signals gracefully
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		if opts.Verbose {
			fmt.Println("[GoBlip] signal received, shutting down")
		}
		w.Stop()
		r.Stop()
		os.Exit(0)
	}()

	if err := w.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		r.Stop()
		os.Exit(1)
	}
}
