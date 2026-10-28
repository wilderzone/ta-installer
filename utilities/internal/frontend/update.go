package frontend

import (
	"context"
	"fmt"
	"strings"
)

// UPDATE is the update utility. It's defined here because play also checks for updates before it starts the game.
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
	return interpretUpdateCheck(i.query(ctx, "check-updates"))
}

// interpretUpdateCheck turns the result of the CLI's `check-updates` subcommand into the update it found, if any.
func interpretUpdateCheck(lines []line, reported []string, err error) (string, error) {
	if len(reported) != 0 {
		return "", fmt.Errorf("check for updates: %s", strings.Join(reported, " "))
	}
	// A cancelled check is passed on as it is, so the caller can tell it apart.
	if err != nil {
		return "", err
	}
	for _, l := range lines {
		switch l.Type {
			case "UPDATE_FOUND":
				return l.Message, nil

			case "INSTALL_UP_TO_DATE":
				return "", nil
		}
	}
	return "", fmt.Errorf("check for updates: no result reported")
}
