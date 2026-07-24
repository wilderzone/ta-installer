package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Counter struct {
	logger  *Logger
	spinner *Spinner
	mu      sync.Mutex
	index   int
	total   int
	label   string
}

// next advances the counter and logs a progress line. It is safe to call from
// multiple goroutines; the lock also serialises the progress output.
func (c *Counter) next(message string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.index++
	progress := PROGRESS
	progress.Index = c.index
	progress.Total = c.total
	c.spinner.update(c.status())
	c.logger.log(progress, message)
}

// status is the text shown beside the spinner while the counter's work is still
// in flight.
func (c *Counter) status() string {
	return fmt.Sprintf("%s...", c.label)
}

// done stops the counter's spinner. Call it once every item has been counted.
func (c *Counter) done() {
	c.spinner.stop()
}

type LogLine struct {
	Index   int    `json:"index,omitempty"`
	Total   int    `json:"total,omitempty"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

type Logger struct {
	machine bool
	silent  bool
	// mu serialises writes to stdout, so spinner frames are never interleaved
	// with log lines. It also guards the spinner's own fields.
	mu      sync.Mutex
	spinner *Spinner
}

// Spinner animates a single line at the bottom of the output, so the user can
// tell that a slow step (a large download, say) is still making progress
// between log lines. A nil *Spinner is a valid, inert spinner: every method
// tolerates it, so callers never need to check whether one was started.
type Spinner struct {
	logger   *Logger
	message  string
	frame    int
	once     sync.Once
	quit     chan struct{}
	finished chan struct{}
}

var (
	ASSET_CREATED             = LogLine{Type: "ASSET_CREATED", Message: ""}
	ASSETS_ARE_DAMAGED        = LogLine{Type: "ASSETS_ARE_DAMAGED", Message: ""}
	CHECKING_ASSETS           = LogLine{Type: "CHECKING_ASSETS", Message: "Checking assets..."}
	CHECKING_FOR_UPDATES      = LogLine{Type: "CHECKING_FOR_UPDATES", Message: "Checking for updates..."}
	DOWNLOADING_MANIFEST      = LogLine{Type: "DOWNLOADING_MANIFEST", Message: "Downloading manifest from:"}
	INSTALL_FAILURE           = LogLine{Type: "INSTALL_FAILURE", Message: "Installation failed."}
	INSTALL_FOUND             = LogLine{Type: "INSTALL_FOUND", Message: "Installation found."}
	INSTALL_NOT_FOUND         = LogLine{Type: "INSTALL_NOT_FOUND", Message: "Installation not found."}
	INSTALL_PARTIAL           = LogLine{Type: "INSTALL_PARTIAL", Message: "Partially installed. You may need to run the \"verify\" command."}
	INSTALL_SUCCESS           = LogLine{Type: "INSTALL_SUCCESS", Message: "Installed."}
	INSTALL_UP_TO_DATE        = LogLine{Type: "INSTALL_UP_TO_DATE", Message: ""}
	INSTALL_VERIFIED          = LogLine{Type: "INSTALL_VERIFIED", Message: "Installation verified."}
	INSTALLING                = LogLine{Type: "INSTALLING", Message: "Installing..."}
	INSTALLING_DEPENDENCIES   = LogLine{Type: "INSTALLING_DEPENDENCIES", Message: "Also installing required assets:"}
	MANIFEST_CONTENTS         = LogLine{Type: "MANIFEST_CONTENTS", Message: ""}
	MANIFEST_CREATED          = LogLine{Type: "MANIFEST_CREATED", Message: ""}
	MANIFEST_UNREADABLE       = LogLine{Type: "MANIFEST_UNREADABLE", Message: ""}
	PATCHING_ASSETS           = LogLine{Type: "PATCHING_ASSETS", Message: "Patching assets..."}
	PREREQUISITES_DOWNLOADING = LogLine{Type: "PREREQUISITES_DOWNLOADING", Message: "Downloading DirectX 9..."}
	PREREQUISITES_FAILED      = LogLine{Type: "PREREQUISITES_FAILED", Message: "Failed to install DirectX 9:"}
	PREREQUISITES_INSTALLING  = LogLine{Type: "PREREQUISITES_INSTALLING", Message: "Installing DirectX 9..."}
	PREREQUISITES_REQUIRED    = LogLine{Type: "PREREQUISITES_REQUIRED", Message: "DirectX 9 needs to be installed before continuing."}
	PROMPT_PROCEED            = LogLine{Type: "PROMPT_PROCEED", Message: "Proceed? [y/N] "}
	PROMPT_UNINSTALL_ANYWAY   = LogLine{Type: "PROMPT_UNINSTALL_ANYWAY", Message: "Remove all files in this folder anyway? [y/N] "}
	READING_LOCAL_MANIFEST    = LogLine{Type: "READING_LOCAL_MANIFEST", Message: "Reading local manifest from:"}
	READING_MANIFEST_FAILED   = LogLine{Type: "READING_MANIFEST_FAILED", Message: "Failed to read manifest:"}
	REDOWNLOADING             = LogLine{Type: "REDOWNLOADING", Message: "Redownloading..."}
	REPO_FOUND                = LogLine{Type: "REPO_FOUND", Message: "Repository found."}
	REPO_NOT_FOUND            = LogLine{Type: "REPO_NOT_FOUND", Message: "Repository not found."}
	SHORTCUTS_CREATING        = LogLine{Type: "SHORTCUTS_CREATING", Message: "Creating shortcuts..."}
	SHORTCUTS_FAILURE         = LogLine{Type: "SHORTCUTS_FAILURE", Message: "Failed to create shortcuts. You may need to re-run the installer with admin permissions to create shortcuts."}
	SHORTCUTS_SUCCESS         = LogLine{Type: "SHORTCUTS_SUCCESS", Message: "Created shortcuts."}
	SHORTCUTS_REMOVE_FAILURE  = LogLine{Type: "SHORTCUTS_REMOVE_FAILURE", Message: "Failed to remove some shortcuts. You may need to delete them manually."}
	UNINSTALL_DEPENDENTS      = LogLine{Type: "UNINSTALL_DEPENDENTS", Message: "Uninstalling the selected assets will also remove these assets:"}
	UNINSTALL_FAILURE         = LogLine{Type: "UNINSTALL_FAILURE", Message: "Failed to uninstall."}
	UNINSTALL_MATCHED_ALL     = LogLine{Type: "UNINSTALL_MATCHED_ALL", Message: "The provided filters match all assets. This will remove the entire installation."}
	UNINSTALL_ORPHANS         = LogLine{Type: "UNINSTALL_ORPHANS", Message: "Also removing assets that are no longer needed:"}
	UNINSTALL_PARTIAL         = LogLine{Type: "UNINSTALL_PARTIAL", Message: "Partially uninstalled. You may need to remove some files manually."}
	UNINSTALL_SUCCESS         = LogLine{Type: "UNINSTALL_SUCCESS", Message: "Uninstalled."}
	UNINSTALLING              = LogLine{Type: "UNINSTALLING", Message: "Uninstalling..."}
	UPDATE_FAILURE            = LogLine{Type: "UPDATE_FAILURE", Message: "Update failed."}
	UPDATE_FOUND              = LogLine{Type: "UPDATE_FOUND", Message: "Update found:"}
	UPDATE_PARTIAL            = LogLine{Type: "UPDATE_PARTIAL", Message: "Partially updated. You may need to run the \"verify\" command."}
	UPDATE_SUCCESS            = LogLine{Type: "UPDATE_SUCCESS", Message: "Updated."}
	UTILITIES_FAILURE         = LogLine{Type: "UTILITIES_FAILURE", Message: "Couldn't download the installation utilities package. Uninstalling the game via the Windows Settings menu or Control Panel may not be possible."}
	UTILITIES_SUCCESS         = LogLine{Type: "UTILITIES_SUCCESS", Message: "Downloaded installation utilities."}
	PROGRESS                  = LogLine{Type: "PROGRESS", Index: 0, Total: 0, Message: ""}
	ERROR                     = LogLine{Type: "ERROR", Message: ""}
)

const (
	RESET      = "\x1b[0m"
	GREY       = "\x1b[2m"
	GREEN      = "\x1b[32m"
	RED        = "\x1b[31m"
	CLEAR_LINE = "\r\x1b[2K" // Return to the start of the line and erase it.
)

var SPINNER_FRAMES = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

const SPINNER_INTERVAL = 100 * time.Millisecond

// isTerminal reports whether stdout is an interactive terminal. The spinner
// redraws itself in place, which is only meaningful there; when stdout is piped
// or redirected the escape codes would just be noise in the captured output.
func isTerminal() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return info.Mode() & os.ModeCharDevice != 0
}

// spin starts a spinner on the line below the output and returns it, so the
// caller can update its message and stop it. It returns nil - an inert spinner -
// when the output isn't an interactive terminal, or in machine and silent modes,
// where redrawing would only corrupt the output.
func (l *Logger) spin(message string) *Spinner {
	if l.machine || l.silent || !isTerminal() {
		return nil
	}
	l.stopSpinner() // Only one spinner can own the line at a time.
	spinner := &Spinner{
		logger:   l,
		message:  message,
		quit:     make(chan struct{}),
		finished: make(chan struct{}),
	}
	l.mu.Lock()
	l.spinner = spinner
	l.mu.Unlock()
	go spinner.run()
	return spinner
}

// run advances the animation until the spinner is stopped.
func (s *Spinner) run() {
	defer close(s.finished)
	ticker := time.NewTicker(SPINNER_INTERVAL)
	defer ticker.Stop()
	for {
		select {
			case <-s.quit:
				return

			case <-ticker.C:
				s.logger.mu.Lock()
				s.frame++
				s.draw()
				s.logger.mu.Unlock()
		}
	}
}

// draw writes the current frame over whatever is on the spinner's line. The
// caller must hold the logger's lock.
func (s *Spinner) draw() {
	text := SPINNER_FRAMES[s.frame % len(SPINNER_FRAMES)]
	if s.message != "" {
		text += " " + s.message
	}
	fmt.Print(CLEAR_LINE + s.logger.colour(text, GREY))
}

// update changes the text shown beside the spinner.
func (s *Spinner) update(message string) {
	if s == nil {
		return
	}
	s.logger.mu.Lock()
	defer s.logger.mu.Unlock()
	s.message = message
}

// stop halts the animation and erases the spinner's line, leaving the cursor
// ready for ordinary output. It is safe to call more than once.
func (s *Spinner) stop() {
	if s == nil {
		return
	}
	s.once.Do(func() {
		close(s.quit)
		<-s.finished // Wait for the last frame, so nothing is drawn after the wipe.
		s.logger.mu.Lock()
		defer s.logger.mu.Unlock()
		if s.logger.spinner == s {
			s.logger.spinner = nil
		}
		fmt.Print(CLEAR_LINE)
	})
}

// stopSpinner halts whichever spinner is currently running, if any. Used before
// output that the spinner must not fight over, such as a prompt.
func (l *Logger) stopSpinner() {
	l.mu.Lock()
	spinner := l.spinner
	l.mu.Unlock()
	spinner.stop()
}

func (l *Logger) colour(message string, colour string) string {
	if l.machine {
		return message
	}
	return colour + message + RESET
}

// count returns a progress counter for `total` items. `label` names the work in
// progress (e.g. "Downloading"), and is shown beside the spinner that runs
// between progress lines. The caller must call done() when the work finishes.
func (l *Logger) count(total int, label string) *Counter {
	counter := &Counter{logger: l, index: 0, total: total, label: label}
	counter.spinner = l.spin(counter.status())
	return counter
}

func (l *Logger) error(message ...any) {
	fmt.Fprintln(os.Stderr, message...)
}

// fail renders a top-level error for the user. It is the single place where
// errors are turned into user-facing output. Errors always go to stderr, keeping
// stdout clean for data and progress logs.
func (l *Logger) fail(err error) {
	l.stopSpinner()

	if l.machine {
		// Structured error line on stderr, so it never corrupts stdout data.
		line := ERROR
		line.Message = err.Error()
		json.NewEncoder(os.Stderr).Encode(line)
		return
	}

	var userError *UserError
	if errors.As(err, &userError) {
		l.error(l.colour(userError.Message, RED))
	} else {
		l.error(l.colour("Something went wrong: ", RED) + err.Error())
	}
}

func (l *Logger) log(message LogLine, additional ...any) {
	if l.silent {
		return
	}

	message.Message = strings.TrimSpace(message.Message + " " + fmt.Sprint(additional...))

	// Hold the lock across the write so a spinner frame can't land mid-line, and
	// wipe the spinner first so it isn't left behind in the scrollback.
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.spinner != nil {
		fmt.Print(CLEAR_LINE)
	}

	if l.machine {
		encoder := json.NewEncoder(os.Stdout)
		err := encoder.Encode(message)
		if err != nil {
			l.error(err.Error())
		}
	} else {
		if message.Type == "PROGRESS" {
			counter := l.colour(fmt.Sprintf("[%d/%d] ", message.Index, message.Total), GREY)
			message.Message = counter + message.Message
		}
		fmt.Println(message.Message)
	}

	// Redraw immediately, so the spinner reappears beneath the new line instead
	// of leaving a gap until the next tick.
	if l.spinner != nil {
		l.spinner.draw()
	}
}

func (l *Logger) plural(count int, singular string, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

func (l *Logger) prompt(message LogLine) bool {
	l.stopSpinner() // The spinner would otherwise redraw over what the user types.

	line := ""
	if l.machine {
		line = message.Type
	} else {
		line = message.Message
	}
	fmt.Print(line)
	action := "n"
	_, _ = fmt.Scanln(&action)
	match, _ := regexp.MatchString("^[yY]", action)
	return match
}
