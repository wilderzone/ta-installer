package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const CLI_EXE = "tai.exe"
const MANIFEST = "manifest.json"

func main() {
	// Always pause before exiting, so the window stays open long enough for the user
	// to read the output - including any error.
	code := run()
	fmt.Print("\nPress Enter to close.")
	_, _ = fmt.Scanln()
	os.Exit(code)
}

// run does the work and returns a process exit code. It never calls os.Exit itself, so
// every path back to main() goes through the pause.
func run() int {
	// Get the path to this executable.
	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	realPath, err := filepath.EvalSymlinks(exePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	directory := filepath.Dir(realPath)

	// Find the TAI CLI executable.
	cli := filepath.Join(directory, CLI_EXE)
	_, err = os.Stat(cli)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	// Find the game's manifest (and thus its root directory).
	game := filepath.Dir(directory)
	manifest := filepath.Join(game, MANIFEST)
	_, err = os.Stat(manifest)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	fmt.Println("Game found.")

	// No %q: exec.Command passes each argument straight to the process without a shell,
	// so quoting would be taken as part of the path.
	command := exec.Command(cli, "uninstall", "-y", "--path=" + game)
	// Stream the CLI's output straight to this console so progress appears live and errors are shown.
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err = command.Run()
	if err != nil {
		return 1
	}

	return 0
}
