package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const MANIFEST = "manifest.json"
const LOG_NAME = "log.txt"
const LOGIN_SERVER = "18.197.240.229"
const LOGS_DIRECTORY = "Logs"
const LAUNCHER_LOG_NAME = "launcher.txt"
const GAME_BINARY_NAME = "TribesAscend.exe"

type arguments struct {
	prepare bool
}

func main() {
	os.Exit(run())
}

// run does the work and returns a process exit code. It never calls os.Exit itself, so
// every path back to main goes through the same reporting.
func run() int {
	args := arguments{}
	fs := flag.NewFlagSet("play", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // Errors are surfaced below instead.
	// Set up the compatibility layer and exit, without starting the game. The installer
	// runs this at the end of an install so the first launch isn't the one that pays for
	// it. On a platform that runs the game directly this does nothing.
	fs.BoolVar(&args.prepare, "prepare", false, "")
	err := fs.Parse(os.Args[1:])
	if err != nil {
		return fail(io.Discard, err)
	}

	game, err := gameDirectory()
	if err != nil {
		return fail(io.Discard, err)
	}

	log, closeLog := openLog(game)
	defer closeLog()

	if args.prepare {
		err = prepareRuntime(game, log)
	} else {
		err = launchGame(game, log)
	}
	if err != nil {
		return fail(log, err)
	}

	return 0
}

// openLog opens the launcher's log, along with the function that closes it.
//
// Everything a launch produces is written here: this launcher's own errors, and on a platform
// with a compatibility layer, whatever umu, Proton and Wine had to say underneath it. A
// shortcut starts the game with no console attached, so without this all of it is lost - and
// it's the only thing worth asking an end user for when the game doesn't start.
//
// Each run replaces the last, because the run being asked about is the one that just happened.
// A log that can't be opened isn't worth failing a launch over, so the caller is handed a
// writer that discards instead.
func openLog(game string) (io.Writer, func()) {
	path := filepath.Join(game, LOGS_DIRECTORY, LAUNCHER_LOG_NAME)
	if os.MkdirAll(filepath.Dir(path), os.FileMode(0755)) != nil {
		return io.Discard, func() {}
	}
	file, err := os.Create(path)
	if err != nil {
		return io.Discard, func() {}
	}

	// Dated, so that a log sent in days later can be told apart from the run it's meant
	// to describe.
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
func gameArguments(logPath string) []string {
	return []string{
		"-hostx=" + LOGIN_SERVER,
		"-abslog=" + logPath,
	}
}

// gameLogPath returns where the game should write its log.
func gameLogPath(game string) string {
	return filepath.Join(game, LOGS_DIRECTORY, LOG_NAME)
}

// fail reports why the game didn't start, to the console if there is one and to the launcher's
// log either way - where it lands after whatever the compatibility layer said on its way down,
// which is usually the part that explains it.
func fail(log io.Writer, err error) int {
	fmt.Fprintln(os.Stderr, err.Error())
	fmt.Fprintln(log, err.Error())
	return 1
}
