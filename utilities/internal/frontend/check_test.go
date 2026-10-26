package frontend

import (
	"errors"
	"testing"
)

func TestInterpretCheckFound(t *testing.T) {
	got, err := interpretCheck([]byte(`{"type":"INSTALL_FOUND","message":"Installation found."}` + "\n"), nil, nil)
	if err != nil || got != CHECK_FOUND {
		t.Errorf("got %q, %v, want %q", got, err, CHECK_FOUND)
	}
}

func TestInterpretCheckNotFound(t *testing.T) {
	got, err := interpretCheck([]byte(`{"type":"INSTALL_NOT_FOUND","message":"Installation not found."}` + "\n"), nil, nil)
	if err != nil || got != CHECK_NOT_FOUND {
		t.Errorf("got %q, %v, want %q", got, err, CHECK_NOT_FOUND)
	}
}

func TestInterpretCheckUnreadable(t *testing.T) {
	stderr := []byte(`{"type":"ERROR","message":"Found an installation at G:\\tribes, but its manifest couldn't be read."}` + "\n")
	got, err := interpretCheck(nil, stderr, errors.New("exit status 1"))
	if err != nil || got != CHECK_UNREADABLE {
		t.Errorf("got %q, %v, want %q", got, err, CHECK_UNREADABLE)
	}
}

func TestInterpretCheckThatCouldNotRun(t *testing.T) {
	// The CLI's own argument errors are plain text, and say nothing about the manifest.
	got, err := interpretCheck(nil, []byte("Unexpected argument \"x\".\n"), errors.New("exit status 1"))
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
