package frontend

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// UPDATE is the update utility. It's defined here because play offers the same update
// before it starts the game.
var UPDATE = Utility{
	Command:     "update",
	Title:       "Update " + PRODUCT_NAME,
	Running:     "Checking for updates...",
	Success:     PRODUCT_NAME + " has been updated.",
	Failure:     "Couldn't update " + PRODUCT_NAME + ".",
	Conclusions: map[string]string{
		"INSTALL_UP_TO_DATE": PRODUCT_NAME + " is already up to date.",
	},
}

// CheckForUpdates runs the CLI's `check-updates` subcommand against the installation. It
// returns the CLI's description of the update it found (eg, "Update found: v12 → v13"), or an
// empty string if the installation is up to date. The check is abandoned once ctx is done.
func (i Installation) CheckForUpdates(ctx context.Context) (string, error) {
	command := exec.CommandContext(ctx, i.CLI, "check-updates", "-m", "--path=" + i.Game)
	hideWindow(command)
	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if ctx.Err() != nil {
		return "", ctx.Err()
	}
	return interpretUpdateCheck(stdout.Bytes(), stderr.Bytes(), err)
}

// interpretUpdateCheck turns the output of the CLI's `check-updates` subcommand into the
// update it found, if any.
func interpretUpdateCheck(stdout []byte, stderr []byte, err error) (string, error) {
	if err != nil {
		for _, l := range parseLines(stderr) {
			if l.Type == "ERROR" {
				return "", fmt.Errorf("check for updates: %s", l.Message)
			}
		}
		return "", fmt.Errorf("check for updates: %w: %s", err, strings.TrimSpace(string(stderr)))
	}
	for _, l := range parseLines(stdout) {
		switch l.Type {
			case "UPDATE_FOUND":
				return l.Message, nil

			case "INSTALL_UP_TO_DATE":
				return "", nil
		}
	}
	return "", fmt.Errorf("check for updates: no result reported")
}
