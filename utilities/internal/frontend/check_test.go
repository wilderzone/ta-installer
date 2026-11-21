package frontend

import (
	"errors"
	"testing"
)

func TestInterpretCheckFound(t *testing.T) {
	got, err := interpretCheck([]line{{Type: "INSTALL_FOUND", Message: "Installation found."}}, nil, nil)
	if err != nil || got != CHECK_FOUND {
		t.Errorf("got %q, %v, want %q", got, err, CHECK_FOUND)
	}
}

func TestInterpretCheckNotFound(t *testing.T) {
	got, err := interpretCheck([]line{{Type: "INSTALL_NOT_FOUND", Message: "Installation not found."}}, nil, nil)
	if err != nil || got != CHECK_NOT_FOUND {
		t.Errorf("got %q, %v, want %q", got, err, CHECK_NOT_FOUND)
	}
}

func TestInterpretCheckUnreadable(t *testing.T) {
	lines := []line{{Type: "MANIFEST_UNREADABLE", Message: "Found an installation at G:\\tribes, but its manifest couldn't be read."}}
	reported := []string{"Couldn't identify the installation's files."}
	got, err := interpretCheck(lines, reported, errors.New("exit status 1"))
	if err != nil || got != CHECK_UNREADABLE {
		t.Errorf("got %q, %v, want %q", got, err, CHECK_UNREADABLE)
	}
}

func TestInterpretCheckWithAnotherReportedError(t *testing.T) {
	// Only the CLI's own line means the manifest is unreadable. Taking any other error for
	// it would offer to delete everything in the installation directory.
	got, err := interpretCheck(nil, []string{"Something went wrong."}, errors.New("exit status 1"))
	if err == nil {
		t.Errorf("got %q, want an error", got)
	}
}

func TestInterpretCheckThatCouldNotRun(t *testing.T) {
	// The CLI's own argument errors are plain text, so they aren't reported errors, and say
	// nothing about the manifest.
	got, err := interpretCheck(nil, nil, errors.New("exit status 1: Unexpected argument \"x\"."))
	if err == nil {
		t.Errorf("got %q, want an error", got)
	}
}

func TestInterpretCheckWithoutAResult(t *testing.T) {
	got, err := interpretCheck(nil, nil, nil)
	if err == nil {
		t.Errorf("got %q, want an error", got)
	}
}

func TestParseLinesSkipsAnythingElse(t *testing.T) {
	output := []byte(`{"type":"INSTALL_FOUND","message":"Installation found."}` + "\nnot json\n")
	got := parseLines(output)
	if len(got) != 1 || got[0].Type != "INSTALL_FOUND" {
		t.Errorf("got %+v, want only the JSON line", got)
	}
}
