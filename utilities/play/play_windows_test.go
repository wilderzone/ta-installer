package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestStartProcessPassesTheArguments(t *testing.T) {
	cmd := filepath.Join(os.Getenv("SystemRoot"), "System32", "cmd.exe")
	// The exit code only comes back if the arguments arrived intact.
	process, err := startProcess(cmd, []string{cmd, "/c", "exit", "3"}, t.TempDir(), "Wilderzone.TribesAscend.Test", false)
	if err != nil {
		t.Fatalf("start process: %v", err)
	}
	state, err := process.Wait()
	if err != nil {
		t.Fatalf("wait for process: %v", err)
	}
	if state.ExitCode() != 3 {
		t.Errorf("got exit code %d, want 3", state.ExitCode())
	}
}

func TestTheGameCantBePinnedWithoutAShortcut(t *testing.T) {
	// A test isn't started from a shortcut, so there's no ID to pass on.
	appID, pinnable := gameAppID(io.Discard)
	if appID != UNPINNABLE_APP_ID || pinnable {
		t.Errorf("got %q, pinnable=%t; want %q, not pinnable", appID, pinnable, UNPINNABLE_APP_ID)
	}
}

func TestShortcutAppIDOfAShortcutWithoutOne(t *testing.T) {
	// A shortcut as written by an installer from before shortcuts had IDs.
	path := filepath.Join(t.TempDir(), "Tribes Ascend.lnk")
	quote := func(value string) string { return "'" + strings.ReplaceAll(value, "'", "''") + "'" }
	script := "$s = (New-Object -ComObject WScript.Shell).CreateShortcut(" + quote(path) + "); $s.TargetPath = " + quote(os.Args[0]) + "; $s.Save()"
	if output, err := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script).CombinedOutput(); err != nil {
		t.Fatalf("create shortcut: %v: %s", err, output)
	}

	appID, err := shortcutAppID(path)
	if err != nil {
		t.Fatalf("read AppUserModelID: %v", err)
	}
	if appID != "" {
		t.Errorf("got AppUserModelID %q, want none", appID)
	}
}

func TestShortcutAppIDOfAMissingShortcut(t *testing.T) {
	if _, err := shortcutAppID(filepath.Join(t.TempDir(), "Missing.lnk")); err == nil {
		t.Error("got no error for a shortcut that doesn't exist")
	}
}
