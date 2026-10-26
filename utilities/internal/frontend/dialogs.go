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

// runWithDialogs runs the CLI behind a progress dialog, then reports how it went.
//
// The CLI is never left to prompt: it is passed `-y`, and anything worth asking is asked
// before it starts. Its standard input is then only ever used to cancel it, so a cancellation
// can't be mistaken for the answer to a prompt, or swallow one.
func runWithDialogs(found installation, utility Utility) int {
	// Open the dialog before starting anything, so a dialog that can't be shown never leaves
	// the CLI working where the user can't see it.
	dialog, err := zenity.Progress(zenity.Title(utility.Title), zenity.MaxValue(PROGRESS_SCALE))
	if err != nil {
		return 1
	}
	view := progress{text: utility.Running}
	_ = dialog.Text(view.text)

	log := openLog(found.game, utility.Command)
	report := func(err error) int {
		log.write(err.Error())
		log.close()
		result{kind: "error", text: utility.Failure + "\n\n" + err.Error()}.show(utility, log.path)
		return 1
	}

	// `-v` adds the details that the CLI otherwise hides, which is exactly what the log is for.
	arguments := []string{utility.Command, "-m", "-p", "-y", "-v", "--path=" + found.game}
	log.write(found.cli + " " + strings.Join(arguments, " "))
	command := exec.Command(found.cli, arguments...)
	hideWindow(command)
	stdin, stdout, stderr, err := pipes(command)
	if err != nil {
		_ = dialog.Close()
		return report(err)
	}
	err = command.Start()
	if err != nil {
		_ = dialog.Close()
		return report(err)
	}

	lines := make(chan line)
	go readOutput(stdout, lines)
	failures := make(chan []string, 1)
	go readErrors(stderr, log, failures)

	outcome := outcome{}
	cancelled := false
	var stopping zenity.ProgressDialog
	closed := dialog.Done()
	for running := true; running; {
		select {
			case l, ok := <-lines:
				if !ok {
					running = false
					break
				}
				log.write(l.readable())
				outcome.observe(l)
				if !cancelled && view.apply(l) {
					_ = dialog.Text(view.text)
					_ = dialog.Value(view.value)
				}

			case <-closed:
				// The user closed the dialog. Ask the CLI to stop, and keep reading until it
				// does, so it can say how to resume.
				closed = nil
				cancelled = true
				log.write("Cancelled by the user.")
				_, _ = io.WriteString(stdin, CANCEL_COMMAND + "\n")
				stopping = stoppingDialog(utility)
		}
	}

	for _, message := range <-failures {
		outcome.fail(message)
	}
	err = command.Wait()
	if err != nil {
		log.write(fmt.Sprintf("%s %s: %s", CLI_NAME, utility.Command, err.Error()))
	}
	if !cancelled {
		_ = dialog.Close()
	}
	if stopping != nil {
		_ = stopping.Close()
	}
	log.close()

	outcome.describe(utility, err).show(utility, log.path)
	if err != nil {
		return 1
	}
	return 0
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
