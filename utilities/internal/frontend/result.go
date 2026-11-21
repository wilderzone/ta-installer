package frontend

import (
	"errors"
	"strings"

	"github.com/ncruces/zenity"
)

const DETAILS_BUTTON = "Show Details"

// FAILURES are the lines that report a run that failed outright. The CLI reports these as
// findings rather than errors, so it still exits cleanly after FAILURES.
var FAILURES = map[string]bool{
	"INSTALL_FAILURE":   true,
	"UNINSTALL_FAILURE": true,
	"UPDATE_FAILURE":    true,
}

// PARTIALS are the lines that report a run that only managed some of its work.
// Like FAILURES, the CLI still exits cleanly after PARTIALS.
var PARTIALS = map[string]bool{
	"INSTALL_PARTIAL":   true,
	"UNINSTALL_PARTIAL": true,
	"UPDATE_PARTIAL":    true,
}

// WARNINGS are the lines that report something the user should know about,
// even though the run as a whole succeeded.
var WARNINGS = map[string]bool{
	"PREREQUISITES_FAILED":      true,
	"RUNTIME_CHECK_FAILED":      true,
	"RUNTIME_CHECK_SKIPPED":     true,
	"RUNTIME_MISSING":           true,
	"RUNTIME_PREPARE_FAILED":    true,
	"SHORTCUTS_FAILURE":         true,
	"SHORTCUTS_REMOVE_FAILURE":  true,
	"UTILITIES_CHECK_SKIPPED":   true,
	"UTILITIES_REPAIR_DEFERRED": true,
}

// outcome gathers what the result dialog needs to know from the CLI's lines.
type outcome struct {
	// The CLI's explanation of how to resume, if the run was interrupted.
	interrupted  string
	// The CLI's explanation of what was left undone, if the run only partly succeeded.
	partial      string
	// The CLI's warning that the drive may not have room for the downloads, if it gave one.
	shortOfSpace string
	failed       bool
	seen         map[string]bool
	errors       []string
	warnings     []string
}

// observe records a line from the CLI's standard output.
func (o *outcome) observe(l line) {
	if o.seen == nil {
		o.seen = map[string]bool{}
	}
	o.seen[l.Type] = true

	if l.Type == "INTERRUPTED" || strings.HasSuffix(l.Type, "_INTERRUPTED") {
		o.interrupted = l.Message
	}
	if FAILURES[l.Type] {
		o.failed = true
	}
	if PARTIALS[l.Type] {
		o.partial = l.Message
	}
	if WARNINGS[l.Type] {
		o.warnings = append(o.warnings, l.Message)
	}
	if l.Type == "INSUFFICIENT_SPACE" {
		o.shortOfSpace = l.Message
	}
}

// fail records an error the CLI reported on its standard error.
func (o *outcome) fail(message string) {
	o.errors = append(o.errors, message)
}

// complete reports whether a run that exited with the given error did all of its work.
// Warnings don't count against it: they're about things beside the work itself.
func (o *outcome) complete(err error) bool {
	return err == nil && !o.failed && o.partial == "" && o.interrupted == ""
}

// exitCode returns the utility's exit code for a run that exited with the given error. The
// CLI exits cleanly from a run that failed or only partly succeeded, but whatever started the
// utility (eg, a quiet uninstall) needs to know that the work wasn't done.
func (o *outcome) exitCode(err error) int {
	if o.complete(err) {
		return 0
	}
	return 1
}

// result is the dialog shown once the CLI has finished.
type result struct {
	kind string // One of "info", "warning", or "error".
	text string
}

// describe decides what to tell the user about a run that exited with the given error.
func (o *outcome) describe(utility Utility, err error) result {
	// An interrupted run exits with an error too, but it stopped safely and said how to
	// pick up where it left off - it isn't a failure.
	if o.interrupted != "" {
		return result{kind: "warning", text: o.interrupted}
	}
	if err != nil || o.failed {
		text := utility.Failure
		if len(o.errors) != 0 {
			text = strings.Join(o.errors, "\n\n")
		}
		return result{kind: "error", text: text}
	}
	// What was left undone replaces the success message, rather than contradicting it.
	if o.partial != "" {
		return result{kind: "warning", text: strings.Join(append([]string{o.partial}, o.warnings...), "\n\n")}
	}

	success := utility.Success
	for kind, text := range utility.Conclusions {
		if o.seen[kind] {
			success = text
		}
	}
	if len(o.warnings) != 0 {
		return result{kind: "warning", text: success + "\n\n" + strings.Join(o.warnings, "\n\n")}
	}
	return result{kind: "info", text: success}
}

// show displays the result, offering to open the run's log when there is one.
func (r result) show(utility Utility, logPath string) {
	options := []zenity.Option{zenity.Title(utility.Title)}
	if logPath != "" {
		// A second button turns a Windows message box into a Yes/No one, and only the
		// labels that are given are changed - so without OKLabel, "OK" would read "Yes".
		options = append(options, zenity.ExtraButton(DETAILS_BUTTON), zenity.OKLabel("OK"))
	}

	var err error
	switch r.kind {
		case "error":
			err = zenity.Error(r.text, options...)

		case "warning":
			err = zenity.Warning(r.text, options...)

		default:
			err = zenity.Info(r.text, options...)
	}

	if errors.Is(err, zenity.ErrExtraButton) {
		_ = openFile(logPath)
	}
}

// ask displays the result as a question, and reports whether the user chose `yes` over `no`.
// It offers to open the run's log as well, and asks again once the log has been opened.
func (r result) ask(utility Utility, logPath string, yes string, no string) bool {
	icon := zenity.InfoIcon
	switch r.kind {
		case "error":
			icon = zenity.ErrorIcon

		case "warning":
			icon = zenity.WarningIcon
	}
	options := []zenity.Option{zenity.Title(utility.Title), zenity.Icon(icon), zenity.OKLabel(yes), zenity.CancelLabel(no)}
	if logPath != "" {
		options = append(options, zenity.ExtraButton(DETAILS_BUTTON))
	}

	for {
		err := zenity.Question(r.text, options...)
		if errors.Is(err, zenity.ErrExtraButton) {
			_ = openFile(logPath)
			continue
		}
		return err == nil
	}
}
