package frontend

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// What the CLI's `check` subcommand can find at an installation.
const (
	CHECK_FOUND      = "found"
	CHECK_NOT_FOUND  = "not found"
	// There's an installation, but its manifest couldn't be read, so there's no telling
	// which of the files around it belong to the game.
	CHECK_UNREADABLE = "unreadable"
)

// Check runs the CLI's `check` subcommand against the installation, and returns what it found.
// An error means the check itself couldn't be run, which says nothing about the installation.
func (i Installation) Check() (string, error) {
	return interpretCheck(i.query(context.Background(), "check"))
}

// interpretCheck turns the result of the CLI's `check` subcommand into a usable result.
//
// An unreadable manifest is the one error `check` reports itself, which tells it apart from the
// CLI failing to run the check at all.
func interpretCheck(lines []line, reported []string, err error) (string, error) {
	if err != nil {
		if len(reported) != 0 {
			return CHECK_UNREADABLE, nil
		}
		return "", fmt.Errorf("check the installation: %w", err)
	}
	for _, l := range lines {
		switch l.Type {
			case "INSTALL_FOUND":
				return CHECK_FOUND, nil

			case "INSTALL_NOT_FOUND":
				return CHECK_NOT_FOUND, nil
		}
	}
	return "", fmt.Errorf("check the installation: no result reported")
}

// query runs a CLI subcommand that only reports on the installation, and returns the lines it
// wrote and the messages of the errors it reported. It's abandoned once ctx is done.
func (i Installation) query(ctx context.Context, subcommand string) ([]line, []string, error) {
	command := exec.CommandContext(ctx, i.CLI, subcommand, "-m", "--path=" + i.Game)
	hideWindow(command)
	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if ctx.Err() != nil {
		return nil, nil, ctx.Err()
	}

	reported := []string{}
	for _, l := range parseLines(stderr.Bytes()) {
		if l.Type == "ERROR" {
			reported = append(reported, l.Message)
		}
	}
	// If nothing was reported and there was no error from the command, use the CLI's error output.
	if err != nil && len(reported) == 0 {
		err = fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return parseLines(stdout.Bytes()), reported, err
}

// parseLines returns the JSON lines in the given output, skipping any lines that aren't JSON-formatted.
func parseLines(output []byte) []line {
	lines := []line{}
	scanner := bufio.NewScanner(bytes.NewReader(output))
	scanner.Buffer(nil, MAX_LINE_LENGTH)
	for scanner.Scan() {
		l := line{}
		if json.Unmarshal(scanner.Bytes(), &l) == nil {
			lines = append(lines, l)
		}
	}
	return lines
}
