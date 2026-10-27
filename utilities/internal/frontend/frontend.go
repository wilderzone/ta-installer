// Package frontend runs the installation utilities: small programs that each run one TAI CLI
// subcommand against the installation they live in, and show the user dialogs instead of a
// terminal window.
package frontend

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ncruces/zenity"
)

const PRODUCT_NAME = "Tribes: Ascend"
const MANIFEST = "manifest.json"
const REOPENED_FLAG = "reopened"

// Utility describes one of the installation utilities: the CLI subcommand it runs, and the words its dialogs use.
type Utility struct {
	// The CLI subcommand to run (eg, "verify").
	Command      string
	// Heads each of the utility's dialogs (eg, "Verify Tribes: Ascend").
	Title        string
	// Shown in the progress dialog until the CLI reports something of its own.
	Running      string
	// Shown when the CLI finishes without error.
	Success      string
	// Shown when the CLI fails without explaining why.
	Failure      string
	// Shown instead of Success when the CLI reports the line of the given type (eg, an
	// update that finds the installation already up to date). A run is expected to report
	// at most one of these.
	Conclusions  map[string]string
	// Keeps the run's log in the temp directory, rather than in the installation, for a
	// utility that removes the installation.
	TemporaryLog bool
	// Asks the user whether to go ahead, before the CLI is started. Returns false to stop.
	// Without dialogs there is no asking this way, so the CLI asks its own questions instead.
	Confirm      func(found Installation) bool
	// Runs once the CLI has finished its work completely: without failing, being
	// interrupted, or leaving any of it undone.
	Complete     func(found Installation)
}

// Installation is where the utility found itself: the game it belongs to, and the copy of the
// CLI that was installed alongside it.
type Installation struct {
	Game string
	CLI  string
}

// Run runs the utility, and returns the process exit code. It never calls os.Exit itself, so
// the utility's main decides how to leave.
func Run(utility Utility) int {
	// zenity draws its own dialogs on Windows, but on Linux it needs a dialog tool to be
	// installed, and there may not be one.
	dialogs := zenity.IsAvailable()

	silent := false
	fs := flag.NewFlagSet(utility.Command, flag.ContinueOnError)
	fs.SetOutput(io.Discard) // Errors are surfaced below instead.
	// Run without showing anything, and without asking: the CLI's `-s` implies `-y` too.
	// Windows runs the uninstaller like this from its QuietUninstallString.
	fs.BoolVar(&silent, "s", false, "")
	// Set on the copy of the utility that reopenInTerminal starts, so it never tries again.
	reopened := false
	fs.BoolVar(&reopened, REOPENED_FLAG, false, "")
	err := fs.Parse(os.Args[1:])
	if err != nil {
		return fail(utility, dialogs && !silent, err.Error())
	}

	found, err := Locate()
	if err != nil {
		message := fmt.Sprintf("Couldn't find the %s installation that this utility belongs to.\n\n%s", PRODUCT_NAME, err.Error())
		return fail(utility, dialogs && !silent, message)
	}

	if silent {
		return runSilently(found, utility)
	}
	if !dialogs {
		// Without dialogs, a terminal is the only way left to show anything, so a utility
		// started without one (eg, from a desktop entry) starts again inside one.
		if reopened || inTerminal() {
			return runInTerminal(found, utility)
		}
		err := reopenInTerminal()
		if err != nil {
			return fail(utility, false, err.Error())
		}
		return 0
	}
	return runWithDialogs(found, utility)
}

// reopenInTerminal starts this executable again, with the same arguments, in a new terminal window.
func reopenInTerminal() error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	return openTerminal(append([]string{executable, "-" + REOPENED_FLAG}, os.Args[1:]...))
}

// fail reports a problem that stopped the utility before the CLI was started, in a dialog if one can be shown.
func fail(utility Utility, dialog bool, message string) int {
	if dialog {
		_ = zenity.Error(message, zenity.Title(utility.Title))
	} else {
		fmt.Fprintln(os.Stderr, message)
	}
	return 1
}

// Locate finds the installation this executable was installed into, confirmed by the manifest
// at its root, and the CLI that sits beside this executable.
func Locate() (Installation, error) {
	executable, err := os.Executable()
	if err != nil {
		return Installation{}, err
	}
	resolved, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return Installation{}, err
	}

	// The game's root is the parent of this utilities directory.
	directory := filepath.Dir(resolved)
	found := Installation{
		Game: filepath.Dir(directory),
		CLI:  filepath.Join(directory, CLI_NAME),
	}

	_, err = os.Stat(filepath.Join(found.Game, MANIFEST))
	if err != nil {
		return Installation{}, fmt.Errorf("no installation found at %s: %w", found.Game, err)
	}
	_, err = os.Stat(found.CLI)
	if err != nil {
		return Installation{}, fmt.Errorf("the installation at %s is missing %s: %w", found.Game, CLI_NAME, err)
	}

	return found, nil
}

// runInTerminal runs the CLI with its output and input connected straight to this process's
// own, then waits for the user to close the window.
//
// This is the fallback for a system with no way to show dialogs, so it assumes it is running
// in a terminal.
func runInTerminal(found Installation, utility Utility) int {
	// No %q: exec.Command passes each argument straight to the process without a shell,
	// so quoting would be taken as part of the path.
	arguments := []string{utility.Command, "--path=" + found.Game}
	// A utility that would have asked first leaves the asking to the CLI.
	if utility.Confirm == nil {
		arguments = append(arguments, "-y")
	}
	command := exec.Command(found.CLI, arguments...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin
	err := command.Run()

	// The CLI's output isn't read here, so its exit code is all there is to go on.
	if err == nil && utility.Complete != nil {
		utility.Complete(found)
	}

	// Pause before exiting, so the window stays open long enough for the user to read the
	// output - including any error.
	fmt.Print("\nPress Enter to close.")
	_, _ = fmt.Scanln()

	if err != nil {
		return 1
	}
	return 0
}
