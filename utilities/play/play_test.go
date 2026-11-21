package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestUnrecognisedArgumentsAreIgnored(t *testing.T) {
	// A shortcut written for a newer launcher may pass flags this one doesn't know, which
	// mustn't stop the game from starting, nor hide the flags it does know.
	for _, test := range []struct {
		name string
		command []string
		loginServer string
		prepare bool
		ignored []string
	}{
		{"no arguments", []string{}, DEFAULT_LOGIN_SERVER, false, []string{}},
		{"only known flags", []string{"-hostx=ta.example.com", "-prepare"}, "ta.example.com", true, []string{}},
		{"two dashes", []string{"--hostx=ta.example.com"}, "ta.example.com", false, []string{}},
		{"unknown flag with its value", []string{"-hostx=ta.example.com", "-appid=Example"}, "ta.example.com", false, []string{"-appid=Example"}},
		{"unknown flag before a known one", []string{"-new", "-prepare"}, DEFAULT_LOGIN_SERVER, true, []string{"-new"}},
		{"unknown flag's value given separately", []string{"-appid", "Example", "-hostx=ta.example.com"}, "ta.example.com", false, []string{"-appid", "Example"}},
		{"known flag's value given separately", []string{"-hostx", "ta.example.com", "-new"}, "ta.example.com", false, []string{"-new"}},
		{"known flag's value that looks like a flag", []string{"-hostx", "-new"}, "-new", false, []string{}},
		{"help flags", []string{"-h", "-help", "-prepare"}, DEFAULT_LOGIN_SERVER, true, []string{"-h", "-help"}},
		{"stray argument", []string{"stray", "-hostx=ta.example.com"}, "ta.example.com", false, []string{"stray"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			args, ignored, err := parseArguments(test.command)
			if err != nil {
				t.Fatalf("parse %q: %v", test.command, err)
			}
			if args.loginServer != test.loginServer || args.prepare != test.prepare {
				t.Errorf("got loginServer=%q, prepare=%t; want loginServer=%q, prepare=%t", args.loginServer, args.prepare, test.loginServer, test.prepare)
			}
			if !slices.Equal(ignored, test.ignored) {
				t.Errorf("ignored %q, want %q", ignored, test.ignored)
			}
		})
	}
}

func TestKnownFlagsWithBadValuesAreStillErrors(t *testing.T) {
	// Ignoring these would start the game with something other than what was asked for.
	for _, command := range [][]string{
		{"-hostx"},
		{"-prepare=maybe"},
	} {
		if _, _, err := parseArguments(command); err == nil {
			t.Errorf("parse %q: got no error", command)
		}
	}
}

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
