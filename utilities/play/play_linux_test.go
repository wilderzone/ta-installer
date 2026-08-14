package main

import (
	"slices"
	"testing"
)

func TestWindowsPathMapsOntoTheZDrive(t *testing.T) {
	got := windowsPath("/home/user/Games/Tribes Ascend/Logs/log.txt")
	if got != `Z:\home\user\Games\Tribes Ascend\Logs\log.txt` {
		t.Errorf("got %s, want a Z: path with backslashes", got)
	}
}

func TestGameArgumentsAreNotQuoted(t *testing.T) {
	// These go straight to the process rather than through a shell or a shortcut's
	// command line, so a quoted path would leave the quotes in the value - and a path
	// split on its spaces would arrive as two arguments.
	got := gameArguments(`Z:\home\user\Games\Tribes Ascend\Logs\log.txt`)
	want := []string{
		"-hostx=" + LOGIN_SERVER,
		`-abslog=Z:\home\user\Games\Tribes Ascend\Logs\log.txt`,
	}
	if !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}
