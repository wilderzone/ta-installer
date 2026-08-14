// Command play starts an installed copy of the game.
//
// It sits beside the other installation utilities and, like them, works out which
// installation it belongs to from its own location. Launching lives here rather than in the
// CLI because TAI is an installer: it puts the game on disk and keeps it healthy, and what
// it takes to actually run the game - the arguments, and on Linux a whole compatibility
// layer - is a separate concern with a separate lifetime.
//
// Keeping it in an executable of its own, rather than in the shortcut, means the launch
// recipe lives in one place: a shortcut written by an older installation keeps working after
// the recipe changes, because an update replaces this binary.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const MANIFEST = "manifest.json"
const LOGS_DIRECTORY = "Logs"
const LOG_NAME = "log.txt"
const LAUNCHER_LOG_NAME = "launcher.txt"
const GAME_BINARY_NAME = "TribesAscend.exe"

// The community login server the game is pointed at.
const LOGIN_SERVER = "18.197.240.229"

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
	// it. On a platform that runs the game directly it does nothing.
	fs.BoolVar(&args.prepare, "prepare", false, "")
	err := fs.Parse(os.Args[1:])
	if err != nil {
		return fail("", err)
	}

	game, err := gameDirectory()
	if err != nil {
		return fail("", err)
	}

	if args.prepare {
		err = prepareRuntime(game)
	} else {
		err = launchGame(game)
	}
	if err != nil {
		return fail(game, err)
	}

	return 0
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

	// The utilities live in <game>/Installer, so the game's root is the grandparent.
	game := filepath.Dir(filepath.Dir(resolved))
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

// fail reports why the game didn't start.
//
// A shortcut may well have started this with no console attached, so the reason is also left
// beside the game's own log, which is somewhere a user can be pointed at.
func fail(game string, err error) int {
	fmt.Fprintln(os.Stderr, err.Error())

	if game != "" {
		path := filepath.Join(game, LOGS_DIRECTORY, LAUNCHER_LOG_NAME)
		if os.MkdirAll(filepath.Dir(path), os.FileMode(0755)) == nil {
			_ = os.WriteFile(path, []byte(err.Error() + "\n"), os.FileMode(0644))
		}
	}

	return 1
}
