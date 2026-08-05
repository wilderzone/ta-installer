package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

const CLI_EXE = "tai-cli.exe"
const MANIFEST = "manifest.json"

type arguments struct {
	silent bool
}

func main() {
	code, pause := run()
	if pause {
		fmt.Print("\nPress Enter to close.")
		_, _ = fmt.Scanln()
	}
	os.Exit(code)
}

// run does the work and returns a process exit code plus whether main should pause before
// closing. It never calls os.Exit itself, so main can always decide about the pause.
func run() (int, bool) {
	// Parse command-line arguments.
	args := arguments{}
	fs := flag.NewFlagSet("uninstall", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // Errors are surfaced to the caller instead.
	fs.BoolVar(&args.silent, "s", false, "")
	err := fs.Parse(os.Args[1:])
	if err != nil {
		return 1, !args.silent
	}

	// Get the path to this executable and the directory it lives in.
	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1, !args.silent
	}
	realPath, err := filepath.EvalSymlinks(exePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1, !args.silent
	}
	directory := filepath.Dir(realPath)

	// Find the TAI CLI executable that sits alongside this one.
	cli := filepath.Join(directory, CLI_EXE)
	_, err = os.Stat(cli)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1, !args.silent
	}

	// The game's root is the parent of this utilities directory.
	game := filepath.Dir(directory)
	manifest := filepath.Join(game, MANIFEST)
	_, err = os.Stat(manifest)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1, !args.silent
	}
	if !args.silent {
		fmt.Println("Game found.")
	}

	// Using `+` instead of `%q`: exec.Command passes each argument straight to the
	// process without a shell, so quoting would be taken as part of the path.
	arguments := []string{
		"uninstall",
		"--path=" + game,
	}
	if args.silent {
		arguments = append(arguments, "-s")
	}
	command := exec.Command(cli, arguments...)
	// Stream the CLI's output straight to this console so progress appears live and errors are shown.
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	// Connect the console's keyboard input too, so the user can answer any prompts the CLI raises.
	command.Stdin = os.Stdin
	err = command.Run()
	if err != nil {
		return 1, !args.silent
	}

	// The CLI removes everything except its own utilities directory, which it can't delete
	// while this executable (and it) are running. Hand that final step to a short-lived
	// batch script that waits for our lock to drop, then removes the whole install directory.
	scheduleCleanup(game)

	return 0, !args.silent
}

