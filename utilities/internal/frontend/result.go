package frontend

import (
	"errors"
	"strings"

	"github.com/ncruces/zenity"
)

const DETAILS_BUTTON = "Show Details"

// WARNINGS are the lines that report something the user should know about, even though the
// run as a whole succeeded.
var WARNINGS = map[string]bool{
	"INSTALL_PARTIAL":           true,
	"PREREQUISITES_FAILED":      true,
	"RUNTIME_CHECK_FAILED":      true,
	"RUNTIME_CHECK_SKIPPED":     true,
	"RUNTIME_MISSING":           true,
	"RUNTIME_PREPARE_FAILED":    true,
	"SHORTCUTS_FAILURE":         true,
	"SHORTCUTS_REMOVE_FAILURE":  true,
	"UNINSTALL_PARTIAL":         true,
	"UPDATE_PARTIAL":            true,
	"UTILITIES_CHECK_SKIPPED":   true,
	"UTILITIES_REPAIR_DEFERRED": true,
}

// outcome gathers what the result dialog needs to know from the CLI's lines.
type outcome struct {
	// The CLI's explanation of how to resume, if the run was interrupted.
	interrupted string
	errors      []string
	warnings    []string
}

// observe records a line from the CLI's standard output.
func (o *outcome) observe(l line) {
	if l.Type == "INTERRUPTED" || strings.HasSuffix(l.Type, "_INTERRUPTED") {
		o.interrupted = l.Message
	}
	if WARNINGS[l.Type] {
		o.warnings = append(o.warnings, l.Message)
	}
}

// fail records an error the CLI reported on its standard error.
func (o *outcome) fail(message string) {
	o.errors = append(o.errors, message)
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
	if err != nil {
		text := utility.Failure
		if len(o.errors) != 0 {
			text = strings.Join(o.errors, "\n\n")
		}
		return result{kind: "error", text: text}
	}
	if len(o.warnings) != 0 {
		return result{kind: "warning", text: utility.Success + "\n\n" + strings.Join(o.warnings, "\n\n")}
	}
	return result{kind: "info", text: utility.Success}
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
