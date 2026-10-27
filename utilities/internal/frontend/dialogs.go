package frontend

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/ncruces/zenity"
)

// CANCEL_COMMAND asks the CLI to stop at its next safe point, when written to its standard
// input. See "Interrupting a run" in the CLI's documentation.
const CANCEL_COMMAND = "cancel"

// MAX_LINE_LENGTH bounds a single line of the CLI's output.
const MAX_LINE_LENGTH = 1024 * 1024

// runWithDialogs runs the CLI behind a progress dialog, then reports the run's result.
func runWithDialogs(found Installation, utility Utility) int {
	if utility.Confirm != nil && !utility.Confirm(found) {
		return 1
	}

	report := RunWithProgress(found, utility)
	report.Show()
	return report.exitCode()
}

// runSilently runs the CLI without showing anything at all. The run is still logged.
// Its output is still read, so that only a run that completely succeeded is followed up.
func runSilently(found Installation, utility Utility) int {
	return execute(found, utility, false).exitCode()
}

// Report is how a run of the CLI went.
type Report struct {
	utility Utility
	outcome outcome
	err     error
	// The run's log, or empty if there isn't one.
	log     string
}

// Complete reports whether the run did all of its work.
func (r Report) Complete() bool {
	return r.outcome.complete(r.err)
}

// exitCode returns the utility's exit code for the run.
func (r Report) exitCode() int {
	return r.outcome.exitCode(r.err)
}

// Warned reports whether the run reported anything the user should know about, even if it completed.
func (r Report) Warned() bool {
	return len(r.outcome.warnings) != 0
}

// Show tells the user how the run went.
func (r Report) Show() {
	r.outcome.describe(r.utility, r.err).show(r.utility, r.log)
}

// Ask tells the user how the run went, and asks them to choose between `yes` and `no`. It reports whether they chose `yes`.
func (r Report) Ask(yes string, no string) bool {
	return r.outcome.describe(r.utility, r.err).ask(r.utility, r.log, yes, no)
}

// RunWithProgress runs the CLI behind a progress dialog, and returns how it went without
// telling the user, so the caller can decide how to.
func RunWithProgress(found Installation, utility Utility) Report {
	return execute(found, utility, true)
}

// execute runs the CLI, behind a progress dialog if it's `visible`, and returns how it went.
// A run that completes its work is followed up with the utility's Complete.
//
// The CLI is never left to prompt: it is passed `-y`, and anything worth asking is asked
// before it starts. Its standard input is then only ever used to cancel an action, so a cancellation
// can't be mistaken for the answer to a prompt, or swallow one.
func execute(found Installation, utility Utility, visible bool) Report {
	report := Report{utility: utility}

	// Open the dialog before starting anything, so a dialog that can't be shown never leaves
	// the CLI working where the user can't see it.
	view := progress{text: utility.Running}
	var dialog zenity.ProgressDialog
	if visible {
		var err error
		dialog, err = zenity.Progress(zenity.Title(utility.Title), zenity.MaxValue(PROGRESS_SCALE))
		if err != nil {
			report.err = err
			return report
		}
		_ = dialog.Text(view.text)
	}

	log := openLog(found, utility)
	report.log = log.path
	// `-v` adds details to the log that the CLI otherwise hides.
	run, err := startCLI(found, []string{utility.Command, "-m", "-p", "-y", "-v", "--path=" + found.Game}, log)
	if err != nil {
		if dialog != nil {
			_ = dialog.Close()
		}
		log.write(err.Error())
		log.close()
		report.err = err
		report.outcome.fail(utility.Failure + "\n\n" + err.Error())
		return report
	}

	outcome := &report.outcome
	cancelled := false
	var stopping zenity.ProgressDialog
	// Without a dialog, there's nothing to close, and this channel never delivers.
	var closed <-chan struct{}
	if dialog != nil {
		closed = dialog.Done()
	}
	for running := true; running; {
		select {
			case l, ok := <-run.lines:
				if !ok {
					running = false
					break
				}
				log.write(l.readable())
				outcome.observe(l)
				if dialog != nil && !cancelled && view.apply(l) {
					_ = dialog.Text(view.text)
					_ = dialog.Value(view.value)
				}

			case <-closed:
				// The user closed the dialog.
				// Ask the CLI to stop, and keep reading until it does, so it can say how to resume.
				closed = nil
				cancelled = true
				log.write("Cancelled by the user.")
				run.cancel()
				stopping = stoppingDialog(utility)
		}
	}

	report.err = run.wait(outcome, utility, log)
	if dialog != nil && !cancelled {
		_ = dialog.Close()
	}
	if stopping != nil {
		_ = stopping.Close()
	}
	log.close()

	if report.Complete() && utility.Complete != nil {
		utility.Complete(found)
	}
	return report
}

// cliRun is the CLI running a subcommand for a utility, with its output read as it arrives.
type cliRun struct {
	command  *exec.Cmd
	stdin    io.WriteCloser
	// Each of the CLI's output lines, closed once its output ends.
	lines    chan line
	// The messages of the errors the CLI reported, sent once its standard error ends.
	failures chan []string
}

// startCLI starts the CLI with the given arguments, recording the run in the log.
func startCLI(found Installation, arguments []string, log *runLog) (*cliRun, error) {
	log.write(found.CLI + " " + strings.Join(arguments, " "))
	command := exec.Command(found.CLI, arguments...)
	hideWindow(command)
	stdin, stdout, stderr, err := pipes(command)
	if err != nil {
		return nil, err
	}
	err = command.Start()
	if err != nil {
		return nil, err
	}

	run := &cliRun{
		command:  command,
		stdin:    stdin,
		lines:    make(chan line),
		failures: make(chan []string, 1),
	}
	go readOutput(stdout, run.lines)
	go readErrors(stderr, log, run.failures)
	return run, nil
}

// cancel asks the CLI to stop at its next safe point.
func (r *cliRun) cancel() {
	_, _ = io.WriteString(r.stdin, CANCEL_COMMAND + "\n")
}

// wait waits for the CLI to exit, once all of its output has been read, and records the
// errors it reported in the outcome.
func (r *cliRun) wait(o *outcome, utility Utility, log *runLog) error {
	for _, message := range <-r.failures {
		o.fail(message)
	}
	err := r.command.Wait()
	if err != nil {
		log.write(fmt.Sprintf("%s %s: %s", CLI_NAME, utility.Command, err.Error()))
	}
	return err
}

// pipes connects to the CLI's standard streams. It must be called before the CLI is started.
func pipes(command *exec.Cmd) (io.WriteCloser, io.ReadCloser, io.ReadCloser, error) {
	stdin, err := command.StdinPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		return nil, nil, nil, err
	}
	return stdin, stdout, stderr, nil
}

// readOutput sends each of the CLI's output lines to `lines`, and closes it once the CLI's
// output ends. A line that isn't JSON is passed on as a message of its own.
func readOutput(stdout io.Reader, lines chan<- line) {
	defer close(lines)
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(nil, MAX_LINE_LENGTH)
	for scanner.Scan() {
		l := line{}
		if json.Unmarshal(scanner.Bytes(), &l) != nil {
			l = line{Message: scanner.Text()}
		}
		lines <- l
	}
}

// readErrors logs each line the CLI writes to its standard error, and sends the messages of
// the errors it reported to `failures` once the stream ends.
//
// The CLI reports its errors as JSON, but its verbose details are plain text, so both are
// expected here.
func readErrors(stderr io.Reader, log *runLog, failures chan<- []string) {
	messages := []string{}
	scanner := bufio.NewScanner(stderr)
	scanner.Buffer(nil, MAX_LINE_LENGTH)
	for scanner.Scan() {
		l := line{}
		if json.Unmarshal(scanner.Bytes(), &l) == nil && l.Type == "ERROR" {
			messages = append(messages, l.Message)
			log.write(l.Message)
			continue
		}
		log.write(scanner.Text())
	}
	failures <- messages
}

// stoppingDialog keeps something on screen while a cancelled CLI works its way to a safe
// point to stop at. It returns nil if the dialog can't be shown.
func stoppingDialog(utility Utility) zenity.ProgressDialog {
	dialog, err := zenity.Progress(zenity.Title(utility.Title), zenity.Pulsate(), zenity.NoCancel())
	if err != nil {
		return nil
	}
	_ = dialog.Text("Stopping...")
	return dialog
}
