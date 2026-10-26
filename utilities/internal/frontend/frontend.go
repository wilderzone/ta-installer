// Package frontend runs the installation utilities: small programs that each run one TAI CLI
// subcommand against the installation they live in, and show the user dialogs instead of a
// terminal window.
package frontend

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ncruces/zenity"
)

const PRODUCT_NAME = "Tribes: Ascend"
const MANIFEST = "manifest.json"

// Utility describes one of the installation utilities: the CLI subcommand it runs, and the
// words its dialogs use.
type Utility struct {
	// The CLI subcommand to run (eg, "verify").
	Command string
	// Heads each of the utility's dialogs (eg, "Verify Tribes: Ascend").
	Title   string
	// Shown in the progress dialog until the CLI reports something of its own.
	Running string
	// Shown when the CLI finishes without error.
	Success string
	// Shown when the CLI fails without explaining why.
	Failure string
}

// installation is where the utility found itself: the game it belongs to, and the copy of the
// CLI that was installed alongside it.
type installation struct {
	game string
	cli  string
}

// Run runs the utility, and returns the process exit code. It never calls os.Exit itself, so
// the utility's main decides how to leave.
func Run(utility Utility) int {
	dialogs := zenity.IsAvailable()

	found, err := locate()
	if err != nil {
		message := fmt.Sprintf("Couldn't find the %s installation that this utility belongs to.\n\n%s", PRODUCT_NAME, err.Error())
		if dialogs {
			_ = zenity.Error(message, zenity.Title(utility.Title))
		} else {
			fmt.Fprintln(os.Stderr, message)
		}
		return 1
	}

	// zenity draws its own dialogs on Windows, but on Linux it needs a dialog tool to be
	// installed, and there may not be one.
	if !dialogs {
		return runInTerminal(found, utility)
	}
	return runWithDialogs(found, utility)
}

// locate finds the installation this executable was installed into, confirmed by the manifest
// at its root, and the CLI that sits beside this executable.
func locate() (installation, error) {
	executable, err := os.Executable()
	if err != nil {
		return installation{}, err
	}
	resolved, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return installation{}, err
	}

	// The game's root is the parent of this utilities directory.
	directory := filepath.Dir(resolved)
	found := installation{
		game: filepath.Dir(directory),
		cli:  filepath.Join(directory, CLI_NAME),
	}

	_, err = os.Stat(filepath.Join(found.game, MANIFEST))
	if err != nil {
		return installation{}, fmt.Errorf("no installation found at %s: %w", found.game, err)
	}
	_, err = os.Stat(found.cli)
	if err != nil {
		return installation{}, fmt.Errorf("the installation at %s is missing %s: %w", found.game, CLI_NAME, err)
	}

	return found, nil
}

// runInTerminal runs the CLI with its output and input connected straight to this process's
// own, then waits for the user to close the window.
//
// This is the fallback for a system with no way to show dialogs, so it assumes it is running
// in a terminal.
func runInTerminal(found installation, utility Utility) int {
	// No %q: exec.Command passes each argument straight to the process without a shell,
	// so quoting would be taken as part of the path.
	command := exec.Command(found.cli, utility.Command, "-y", "--path=" + found.game)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	err := command.Run()

	// Pause before exiting, so the window stays open long enough for the user to read the
	// output - including any error.
	fmt.Print("\nPress Enter to close.")
	_, _ = fmt.Scanln()

	if err != nil {
		return 1
	}
	return 0
}
