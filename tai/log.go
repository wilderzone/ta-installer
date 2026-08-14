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
	logger   *Logger
	spinner  *Spinner
	progress *Progress
	mu       sync.Mutex
	index    int
	total    int
	label    string
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
// in flight. A measured download adds how far through it is, so the line moves
// between the counter's own, much rarer, steps.
func (c *Counter) status() string {
	status := fmt.Sprintf("%s...", c.label)
	if summary := c.progress.summary(); summary != "" {
		status += " " + summary
	}
	return status
}

// done stops the counter's spinner and closes off any measured download. Call it
// once every item has been counted.
func (c *Counter) done() {
	c.progress.stop()
	c.spinner.stop()
}

type LogLine struct {
	Index   int    `json:"index,omitempty"`
	Total   int    `json:"total,omitempty"`
	// Set on the lines a measured download reports itself with: the bytes received so
	// far, the bytes expected in total, and the smoothed transfer rate per second.
	Bytes   uint64 `json:"bytes,omitempty"`
	Size    uint64 `json:"size,omitempty"`
	Rate    uint64 `json:"rate,omitempty"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

type Logger struct {
	machine  bool
	// progress emits the download-progress lines that logProgress describes. They are
	// opt-in because only an interface reading the output has any use for them.
	progress bool
	silent   bool
	verbose  bool
	// mu serialises writes to stdout, so spinner frames are never interleaved
	// with log lines. It also guards the spinner's own fields.
	mu       sync.Mutex
	spinner  *Spinner
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
	DOWNLOAD_COMPLETE         = LogLine{Type: "DOWNLOAD_COMPLETE", Message: ""}
	DOWNLOAD_PROGRESS         = LogLine{Type: "DOWNLOAD_PROGRESS", Message: ""}
	DOWNLOAD_STARTED          = LogLine{Type: "DOWNLOAD_STARTED", Message: ""}
	DOWNLOADING_MANIFEST      = LogLine{Type: "DOWNLOADING_MANIFEST", Message: "Downloading manifest from:"}
	INSTALL_FAILURE           = LogLine{Type: "INSTALL_FAILURE", Message: "Installation failed."}
	INSTALL_FOUND             = LogLine{Type: "INSTALL_FOUND", Message: "Installation found."}
	INSTALL_INTERRUPTED       = LogLine{Type: "INSTALL_INTERRUPTED", Message: "Installation interrupted. Run \"install\" or \"verify\" again to finish installing."}
	INSTALL_NOT_FOUND         = LogLine{Type: "INSTALL_NOT_FOUND", Message: "Installation not found."}
	INSTALL_PARTIAL           = LogLine{Type: "INSTALL_PARTIAL", Message: "Partially installed. You may need to run the \"verify\" command."}
	INSTALL_SUCCESS           = LogLine{Type: "INSTALL_SUCCESS", Message: "Installed."}
	INSTALL_UP_TO_DATE        = LogLine{Type: "INSTALL_UP_TO_DATE", Message: ""}
	INSTALL_VERIFIED          = LogLine{Type: "INSTALL_VERIFIED", Message: "Installation verified."}
	INSTALLING                = LogLine{Type: "INSTALLING", Message: "Installing..."}
	INSTALLING_DEPENDENCIES   = LogLine{Type: "INSTALLING_DEPENDENCIES", Message: "Also installing required assets:"}
	INTERRUPTED               = LogLine{Type: "INTERRUPTED", Message: "Interrupted."}
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
	RUNTIME_CHECK_FAILED      = LogLine{Type: "RUNTIME_CHECK_FAILED", Message: "The compatibility layer may not run on this system:"}
	RUNTIME_CHECK_SKIPPED     = LogLine{Type: "RUNTIME_CHECK_SKIPPED", Message: "Couldn't check the compatibility layer:"}
	RUNTIME_DAMAGED           = LogLine{Type: "RUNTIME_DAMAGED", Message: "The compatibility layer is damaged:"}
	RUNTIME_DOWNLOADING       = LogLine{Type: "RUNTIME_DOWNLOADING", Message: "Downloading the compatibility layer..."}
	RUNTIME_MISSING           = LogLine{Type: "RUNTIME_MISSING", Message: "The compatibility layer is missing. Run \"install\" or \"verify\" to restore it."}
	RUNTIME_PREPARE_FAILED    = LogLine{Type: "RUNTIME_PREPARE_FAILED", Message: "Couldn't prepare the compatibility layer. It will be prepared the first time the game is started."}
	RUNTIME_PREPARING         = LogLine{Type: "RUNTIME_PREPARING", Message: "Preparing the compatibility layer..."}
	RUNTIME_READY             = LogLine{Type: "RUNTIME_READY", Message: "Compatibility layer ready."}
	RUNTIME_REPAIRED          = LogLine{Type: "RUNTIME_REPAIRED", Message: "Replaced the compatibility layer."}
	RUNTIME_VERIFIED          = LogLine{Type: "RUNTIME_VERIFIED", Message: "Compatibility layer verified."}
	SHORTCUTS_FAILURE         = LogLine{Type: "SHORTCUTS_FAILURE", Message: "Failed to create shortcuts. " + SHORTCUTS_FAILURE_HINT}
	SHORTCUTS_SUCCESS         = LogLine{Type: "SHORTCUTS_SUCCESS", Message: "Created shortcuts."}
	SHORTCUTS_REMOVE_FAILURE  = LogLine{Type: "SHORTCUTS_REMOVE_FAILURE", Message: "Failed to remove some shortcuts. You may need to delete them manually."}
	UNINSTALL_DEPENDENTS      = LogLine{Type: "UNINSTALL_DEPENDENTS", Message: "Uninstalling the selected assets will also remove these assets:"}
	UNINSTALL_EVERYTHING      = LogLine{Type: "UNINSTALL_EVERYTHING", Message: fmt.Sprintf("Are you sure you want to uninstall %s?", PRODUCT_NAME)}
	UNINSTALL_FAILURE         = LogLine{Type: "UNINSTALL_FAILURE", Message: "Failed to uninstall."}
	UNINSTALL_INTERRUPTED     = LogLine{Type: "UNINSTALL_INTERRUPTED", Message: "Uninstall interrupted. Run \"uninstall\" again to remove the remaining files."}
	UNINSTALL_MATCHED_ALL     = LogLine{Type: "UNINSTALL_MATCHED_ALL", Message: "The provided filters match all assets. This will remove the entire installation."}
	UNINSTALL_ORPHANS         = LogLine{Type: "UNINSTALL_ORPHANS", Message: "Also removing assets that are no longer needed:"}
	UNINSTALL_PARTIAL         = LogLine{Type: "UNINSTALL_PARTIAL", Message: "Partially uninstalled. You may need to manually remove some files from:"}
	UNINSTALL_SUCCESS         = LogLine{Type: "UNINSTALL_SUCCESS", Message: "Uninstalled."}
	UNINSTALLING              = LogLine{Type: "UNINSTALLING", Message: "Uninstalling..."}
	UPDATE_FAILURE            = LogLine{Type: "UPDATE_FAILURE", Message: "Update failed."}
	UPDATE_FOUND              = LogLine{Type: "UPDATE_FOUND", Message: "Update found:"}
	UPDATE_INTERRUPTED        = LogLine{Type: "UPDATE_INTERRUPTED", Message: "Update interrupted. Run \"update\" or \"verify\" again to finish updating."}
	UPDATE_PARTIAL            = LogLine{Type: "UPDATE_PARTIAL", Message: "Partially updated. You may need to run the \"verify\" command."}
	UPDATE_SUCCESS            = LogLine{Type: "UPDATE_SUCCESS", Message: "Updated."}
	UTILITIES_CHECK_SKIPPED   = LogLine{Type: "UTILITIES_CHECK_SKIPPED", Message: "Couldn't check the installation utilities:"}
	UTILITIES_DAMAGED         = LogLine{Type: "UTILITIES_DAMAGED", Message: "Installation utilities are damaged:"}
	UTILITIES_FAILURE         = LogLine{Type: "UTILITIES_FAILURE", Message: "Couldn't setup the installation utilities package. Uninstalling the game via the Windows Settings menu or Control Panel may not be possible."}
	UTILITIES_REPAIR_DEFERRED = LogLine{Type: "UTILITIES_REPAIR_DEFERRED", Message: "These utilities are in use and will be replaced after this window is closed:"}
	UTILITIES_REPAIRED        = LogLine{Type: "UTILITIES_REPAIRED", Message: "Repaired installation utilities:"}
	UTILITIES_SUCCESS         = LogLine{Type: "UTILITIES_SUCCESS", Message: "Downloaded installation utilities."}
	UTILITIES_VERIFIED        = LogLine{Type: "UTILITIES_VERIFIED", Message: "Installation utilities verified."}
	VERIFY_INTERRUPTED        = LogLine{Type: "VERIFY_INTERRUPTED", Message: "Verification interrupted. Run \"verify\" again to finish checking the installation."}
	PROGRESS                  = LogLine{Type: "PROGRESS", Index: 0, Total: 0, Message: ""}
	ERROR                     = LogLine{Type: "ERROR", Message: ""}
)

// interruption returns the line to log when the given subcommand is cancelled part-way
// through. Each one names the command to run to pick up where the cancelled one left off,
// so an interrupted run is never a dead end.
func interruption(method string) LogLine {
	switch method {
		case "install":
			return INSTALL_INTERRUPTED

		case "uninstall":
			return UNINSTALL_INTERRUPTED

		case "update":
			return UPDATE_INTERRUPTED

		case "verify":
			return VERIFY_INTERRUPTED
	}
	return INTERRUPTED
}

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

// detail surfaces an otherwise-swallowed error on stderr, but only in verbose (-v) mode.
// The regular codepaths deliberately hide these (eg, a failed best-effort step) so this
// is the escape hatch when debugging, without cluttering normal output. A nil error is
// tolerated, so a best-effort call can be wrapped directly at the point it's ignored:
//
//	logger.detail("add uninstaller to registry", addUninstallerToRegistry(path))
func (l *Logger) detail(context string, err error) {
	if err == nil || !l.verbose {
		return
	}
	l.stopSpinner() // Keep the diagnostic line from fighting the spinner.
	l.error(l.colour(context + ": " + err.Error(), GREY))
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

	// In verbose mode, reveal the technical cause that the friendly message hides.
	if l.verbose {
		if cause := errors.Unwrap(err); cause != nil {
			l.error(l.colour("cause: " + cause.Error(), GREY))
		}
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

// logProgress writes a line describing a download in flight. Unlike the ordinary log
// lines, this one is always JSON, in every mode: it is emitted several times a second for
// an interface to read, and a person watching a terminal is shown the spinner instead.
func (l *Logger) logProgress(message LogLine) {
	if l.silent {
		return
	}

	// Hold the lock across the write for the same reason log does, and wipe the spinner
	// first in case one is running alongside the reports.
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.spinner != nil {
		fmt.Print(CLEAR_LINE)
	}

	encoder := json.NewEncoder(os.Stdout)
	err := encoder.Encode(message)
	if err != nil {
		l.error(err.Error())
	}

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
