package frontend

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const LOGS_DIRECTORY = "Logs"

// runLog keeps a readable record of a run, for the result dialog's details button. It is
// replaced on each run, so it only ever describes the one the user is asking about.
type runLog struct {
	// Empty if the log couldn't be created, in which case there are no details to show.
	path   string
	mu     sync.Mutex
	writer io.Writer
	file   *os.File
}

// openLog creates the log for a run of the given CLI subcommand. A log that can't be created
// discards what is written to it instead: nothing about a run depends on its log.
func openLog(game string, command string) *runLog {
	path := filepath.Join(game, LOGS_DIRECTORY, command + ".txt")
	if os.MkdirAll(filepath.Dir(path), os.FileMode(0755)) != nil {
		return &runLog{writer: io.Discard}
	}
	file, err := os.Create(path)
	if err != nil {
		return &runLog{writer: io.Discard}
	}

	fmt.Fprintf(file, "%s\n", time.Now().Format(time.RFC1123))

	return &runLog{path: path, writer: file, file: file}
}

// write adds a line to the log. It is safe to call from multiple goroutines, so the CLI's
// standard output and standard error can be recorded as they arrive.
func (l *runLog) write(text string) {
	if text == "" {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	fmt.Fprintln(l.writer, text)
}

// close finishes the log, so the file is complete before anything opens it.
func (l *runLog) close() {
	if l.file != nil {
		_ = l.file.Close()
	}
}
