package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenLogReplacesThePreviousRun(t *testing.T) {
	game := t.TempDir()

	log, closeLog := openLog(game)
	log.Write([]byte("first run\n"))
	closeLog()

	log, closeLog = openLog(game)
	log.Write([]byte("second run\n"))
	closeLog()

	contents, err := os.ReadFile(filepath.Join(game, LOGS_DIRECTORY, LAUNCHER_LOG_NAME))
	if err != nil {
		t.Fatalf("read the launcher log: %v", err)
	}
	if strings.Contains(string(contents), "first run") {
		t.Error("the log kept the previous run, which makes the one being asked about ambiguous")
	}
	if !strings.Contains(string(contents), "second run") {
		t.Errorf("the log is missing the run that wrote it: %q", contents)
	}
}

func TestOpenLogSurvivesAnUnwritableInstallation(t *testing.T) {
	// A file where the logs directory would go, so it can't be created. Nothing about a
	// launch depends on the log, so this has to be survivable rather than fatal.
	game := filepath.Join(t.TempDir(), "game")
	err := os.WriteFile(game, []byte{}, os.FileMode(0644))
	if err != nil {
		t.Fatalf("create the blocking file: %v", err)
	}

	log, closeLog := openLog(game)
	defer closeLog()
	if log == nil {
		t.Fatal("got no writer, want one that discards")
	}
	_, err = log.Write([]byte("ignored\n"))
	if err != nil {
		t.Errorf("writing to the fallback log failed: %v", err)
	}
}
