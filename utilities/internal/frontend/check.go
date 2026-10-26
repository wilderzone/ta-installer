package frontend

import (
	"bufio"
	"bytes"
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
	command := exec.Command(i.CLI, "check", "-m", "--path=" + i.Game)
	hideWindow(command)
	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	return interpretCheck(stdout.Bytes(), stderr.Bytes(), err)
}

// interpretCheck turns the output of the CLI's `check` subcommand into what it found.
//
// An unreadable manifest is the one error `check` reports itself. It does so as a JSON line on
// standard error, which tells it apart from the CLI failing to run the check at all.
func interpretCheck(stdout []byte, stderr []byte, err error) (string, error) {
	for _, l := range parseLines(stdout) {
		switch l.Type {
			case "INSTALL_FOUND":
				if err == nil {
					return CHECK_FOUND, nil
				}

			case "INSTALL_NOT_FOUND":
				if err == nil {
					return CHECK_NOT_FOUND, nil
				}
		}
	}
	if err != nil {
		for _, l := range parseLines(stderr) {
			if l.Type == "ERROR" {
				return CHECK_UNREADABLE, nil
			}
		}
		return "", fmt.Errorf("check the installation: %w: %s", err, strings.TrimSpace(string(stderr)))
	}
	return "", fmt.Errorf("check the installation: no result reported")
}

// parseLines returns the JSON lines in the given output, skipping any that aren't.
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
