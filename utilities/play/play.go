package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// The address of the Community login server (18.197.240.229).
const DEFAULT_LOGIN_SERVER = "ta.kfk4ever.com"
const MANIFEST = "manifest.json"
const LOG_NAME = "log.txt"
const LOGS_DIRECTORY = "Logs"
const LAUNCHER_LOG_NAME = "launcher.txt"
const GAME_BINARY_NAME = "TribesAscend.exe"

type arguments struct {
	prepare bool
	loginServer string
}

func main() {
	os.Exit(run())
}

// run does the work and returns a process exit code. It never calls os.Exit itself, so
// every path back to main goes through the same reporting.
func run() int {
	args, ignored, err := parseArguments(os.Args[1:])
	if err != nil {
		return fail(io.Discard, err)
	}

	game, err := gameDirectory()
	if err != nil {
		return fail(io.Discard, err)
	}

	log, closeLog := openLog(game)
	defer closeLog()

	// Log the arguments passed to `play`.
	fmt.Fprintf(log, "Utility arguments: loginServer=%s, prepare=%t\n", args.loginServer, args.prepare)
	if len(ignored) > 0 {
		fmt.Fprintf(log, "Ignored unrecognised arguments: %q\n", ignored)
	}

	if args.prepare {
		err = prepareRuntime(game, log)
	} else {
		if !offerUpdate(log) {
			return 1
		}
		err = launchGame(game, args.loginServer, log)
	}
	if err != nil {
		return fail(log, err)
	}

	return 0
}

// parseArguments reads the launcher's arguments, and returns them along with any it doesn't
// recognise. A recognised flag with a bad value produces an error.
//
// The shortcuts that start this launcher are written by the installer, but the launcher
// itself comes from the default repository, so the two aren't always from the same release.
// A shortcut written for a newer launcher may pass it a flag it doesn't know yet, which
// shouldn't stop the game from starting.
func parseArguments(command []string) (arguments, []string, error) {
	args := arguments{}
	fs := flag.NewFlagSet("play", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // Errors are surfaced by the caller instead.
	// Set up the compatibility layer and exit, without starting the game. The installer
	// runs this at the end of an install so the first launch isn't the one that pays for
	// it. On a platform that runs the game directly this does nothing.
	fs.BoolVar(&args.prepare, "prepare", false, "")
	fs.StringVar(&args.loginServer, "hostx", DEFAULT_LOGIN_SERVER, "")

	ignored := []string{}
	remaining := command
	for {
		err := fs.Parse(remaining)
		// The flag package has no error of its own for an unknown flag, only this message.
		// It has already moved past the flag, so the arguments still to parse are the ones
		// after it.
		if err != nil && (strings.HasPrefix(err.Error(), "flag provided but not defined") || errors.Is(err, flag.ErrHelp)) {
			ignored = append(ignored, remaining[len(remaining) - len(fs.Args()) - 1])
			remaining = fs.Args()
		} else if err == nil && fs.NArg() > 0 {
			// Parsing stops at the first argument that isn't a flag. This launcher takes no
			// positional arguments, so it's most likely the value of an unknown flag before it,
			// and is skipped to get to any flags after it.
			ignored = append(ignored, fs.Arg(0))
			remaining = fs.Args()[1:]
		} else {
			return args, ignored, err
		}
	}
}

// openLog opens the launcher's log, and returns a handle to the file along with the function that closes it.
//
// The log includes this launcher's own errors, and any errors from the compatibility layer.
// Log output is discarded if the log file can't be created or opened.
func openLog(game string) (io.Writer, func()) {
	path := filepath.Join(game, LOGS_DIRECTORY, LAUNCHER_LOG_NAME)
	if os.MkdirAll(filepath.Dir(path), os.FileMode(0755)) != nil {
		return io.Discard, func() {}
	}
	file, err := os.Create(path)
	if err != nil {
		return io.Discard, func() {}
	}

	fmt.Fprintf(file, "%s\n", time.Now().Format(time.RFC1123))

	return file, func() { file.Close() }
}

// gameDirectory returns the installation this executable was installed into, confirmed by
// the manifest sitting at its root.
func gameDirectory() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return "", err
	}

	// The game's root is the parent of this utilities directory.
	directory := filepath.Dir(resolved)
	game := filepath.Dir(directory)
	_, err = os.Stat(filepath.Join(game, MANIFEST))
	if err != nil {
		return "", fmt.Errorf("no installation found at %s: %w", game, err)
	}

	return game, nil
}

// gameArguments returns the argv the game is always started with.
//
// Nothing here is quoted. These are passed straight to the process rather than through a
// shell or a shortcut's command line, so a quote would become part of the value and the game
// would go looking for a log path that includes it.
func gameArguments(loginServer string) []string {
	return []string{
		"-hostx=" + loginServer,
	}
}

// gameLogPath returns where the game should write its log.
func gameLogPath(game string) string {
	return filepath.Join(game, LOGS_DIRECTORY, LOG_NAME)
}

// fail reports why the game didn't start, to the launcher's log file, and to the console (if there is one).
func fail(log io.Writer, err error) int {
	fmt.Fprintln(os.Stderr, err.Error())
	fmt.Fprintln(log, err.Error())
	return 1
}
