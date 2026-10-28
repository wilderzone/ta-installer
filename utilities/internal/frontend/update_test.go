package frontend

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestInterpretUpdateCheckFound(t *testing.T) {
	lines := []line{
		{Type: "CHECKING_FOR_UPDATES", Message: "Checking for updates..."},
		{Type: "UPDATE_FOUND", Message: "Update found: v12 → v13"},
	}
	got, err := interpretUpdateCheck(lines, nil, nil)
	if err != nil || got != "Update found: v12 → v13" {
		t.Errorf("got %q, %v, want the CLI's description of the update", got, err)
	}
}

func TestInterpretUpdateCheckUpToDate(t *testing.T) {
	got, err := interpretUpdateCheck([]line{{Type: "INSTALL_UP_TO_DATE", Message: "Installation is up to date: v13"}}, nil, nil)
	if err != nil || got != "" {
		t.Errorf("got %q, %v, want no update", got, err)
	}
}

func TestInterpretUpdateCheckThatFailed(t *testing.T) {
	reported := []string{"Couldn't check for updates from https://cdn.wilderzone.org/ta/game."}
	_, err := interpretUpdateCheck(nil, reported, errors.New("exit status 1"))
	if err == nil || !strings.Contains(err.Error(), "Couldn't check for updates") {
		t.Errorf("got %v, want the CLI's own error", err)
	}
}

func TestInterpretUpdateCheckThatWasAbandoned(t *testing.T) {
	// play tells a check that took too long apart from one that failed.
	_, err := interpretUpdateCheck(nil, nil, context.DeadlineExceeded)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("got %v, want the context's error", err)
	}
}

func TestInterpretUpdateCheckWithoutAResult(t *testing.T) {
	_, err := interpretUpdateCheck(nil, nil, nil)
	if err == nil {
		t.Error("got no error for a check that reported nothing")
	}
}
