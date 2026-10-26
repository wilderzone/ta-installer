package frontend

import (
	"errors"
	"strings"
	"testing"
)

func TestInterpretUpdateCheckFound(t *testing.T) {
	stdout := []byte(`{"type":"CHECKING_FOR_UPDATES","message":"Checking for updates..."}` + "\n" + `{"type":"UPDATE_FOUND","message":"Update found: v12 → v13"}` + "\n")
	got, err := interpretUpdateCheck(stdout, nil, nil)
	if err != nil || got != "Update found: v12 → v13" {
		t.Errorf("got %q, %v, want the CLI's description of the update", got, err)
	}
}

func TestInterpretUpdateCheckUpToDate(t *testing.T) {
	stdout := []byte(`{"type":"INSTALL_UP_TO_DATE","message":"Installation is up to date: v13"}` + "\n")
	got, err := interpretUpdateCheck(stdout, nil, nil)
	if err != nil || got != "" {
		t.Errorf("got %q, %v, want no update", got, err)
	}
}

func TestInterpretUpdateCheckThatFailed(t *testing.T) {
	stderr := []byte(`{"type":"ERROR","message":"Couldn't check for updates from https://cdn.wilderzone.org/ta/game."}` + "\n")
	_, err := interpretUpdateCheck(nil, stderr, errors.New("exit status 1"))
	if err == nil || !strings.Contains(err.Error(), "Couldn't check for updates") {
		t.Errorf("got %v, want the CLI's own error", err)
	}
}

func TestInterpretUpdateCheckWithoutAResult(t *testing.T) {
	_, err := interpretUpdateCheck(nil, nil, nil)
	if err == nil {
		t.Error("got no error for a check that reported nothing")
	}
}
