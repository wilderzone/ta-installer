package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const RUNTIME_DIRECTORY = "Runtime"
const PREFIX_DIRECTORY = "Prefix"
const UMU_DIRECTORY = "umu"
const UMU_EXECUTABLE = "umu-run"

// The identifier umu uses to look up per-game fixes. Tribes: Ascend's Steam application ID
// opts the standalone build into the same Proton fixes the Steam release gets. If those
// fixes ever do something the standalone build dislikes, "umu-default" turns them off.
const UMU_GAME_ID = "umu-17080"

func runtimeDirectory(game string) string {
	return filepath.Join(game, RUNTIME_DIRECTORY)
}

func prefixDirectory(game string) string {
	return filepath.Join(game, PREFIX_DIRECTORY)
}

func umuExecutable(game string) string {
	return filepath.Join(runtimeDirectory(game), UMU_DIRECTORY, UMU_EXECUTABLE)
}

// protonDirectory finds the unpacked Proton build inside the runtime directory. The build is
// located by looking for the "proton" entry point rather than by name, so that publishing a
// newer build doesn't need a matching change here.
func protonDirectory(game string) (string, error) {
	entries, err := os.ReadDir(runtimeDirectory(game))
	if err != nil {
		return "", fmt.Errorf("read the runtime directory: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		candidate := filepath.Join(runtimeDirectory(game), entry.Name())
		if _, err := os.Stat(filepath.Join(candidate, "proton")); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("no Proton build found in %s", runtimeDirectory(game))
}

// windowsPath converts an absolute Unix path to the form the game sees from inside the
// prefix. Wine maps the whole filesystem onto the Z: drive, so no prefix surgery is needed
// to make an installation outside the prefix reachable.
func windowsPath(path string) string {
	return "Z:" + strings.ReplaceAll(path, "/", `\`)
}

// runtimeCommand builds a command that runs a Windows executable inside this installation's
// prefix. Everything umu needs is passed through the environment; nothing is written to the
// user's shell configuration or to a wrapper script.
//
// The command is wired up to report as it goes, rather than to have its output collected and
// inspected afterwards: umu fetching its container image and Proton setting up a prefix both
// take long enough that their progress is worth seeing, and either can fail in a way only
// their own output explains.
func runtimeCommand(game string, log io.Writer, executable string, args ...string) (*exec.Cmd, error) {
	proton, err := protonDirectory(game)
	if err != nil {
		return nil, err
	}

	environment := []string{
		"GAMEID=" + UMU_GAME_ID,
		"STORE=none",
		"PROTONPATH=" + proton,
		"WINEPREFIX=" + prefixDirectory(game),
		"PROTON_VERB=waitforexitandrun",
		// The runtime is pinned by the installer and verified against published
		// checksums, so leave it alone rather than letting umu fetch a newer one.
		"UMU_RUNTIME_UPDATE=0",
	}

	command := exec.Command(umuExecutable(game), append([]string{executable}, args...)...)
	command.Env = append(os.Environ(), environment...)

	// Which Proton build ran, against which prefix, is the first thing anyone reading this
	// log will need, and none of it can be recovered from the command line alone.
	fmt.Fprintln(log, strings.Join(command.Args, " "))
	for _, variable := range environment {
		fmt.Fprintf(log, "\t%s\n", variable)
	}

	// Written to both, so that a launch from a terminal still shows what it's doing while a
	// launch from a shortcut - which has no console at all - still leaves a record.
	command.Stdout = io.MultiWriter(os.Stdout, log)
	command.Stderr = io.MultiWriter(os.Stderr, log)

	return command, nil
}

// prepareRuntime creates the prefix if it isn't there yet.
//
// This is also what makes umu fetch its container image, which it keeps in the user's data
// directory rather than in the installation - several hundred megabytes that would otherwise
// be paid for on the first launch, with nothing on screen to explain the wait. The installer
// runs this at the end of an install for exactly that reason, but launching falls back to it
// so that an installation whose utilities arrived late still works.
func prepareRuntime(game string, log io.Writer) error {
	if _, err := os.Stat(filepath.Join(prefixDirectory(game), "pfx", "drive_c", "windows")); err == nil {
		return nil
	}

	err := os.MkdirAll(prefixDirectory(game), os.FileMode(0755))
	if err != nil {
		return fmt.Errorf("create the prefix directory: %w", err)
	}

	// wineboot only initialises the prefix and exits, which is all this needs; the point
	// is the setup umu does around it.
	command, err := runtimeCommand(game, log, "wineboot", "-u")
	if err != nil {
		return err
	}
	err = command.Run()
	if err != nil {
		// Whatever went wrong said so on its way past, into this same log and onto the
		// console the installer reads, so there's nothing to repeat here.
		return fmt.Errorf("prepare the compatibility layer: %w", err)
	}

	return nil
}

// launchGame starts the game inside the compatibility layer.
//
// Unlike the Windows launch this waits for the game to exit, because umu is the parent of
// the game's process tree and returning would tear it down.
func launchGame(game string, log io.Writer) error {
	err := prepareRuntime(game, log)
	if err != nil {
		return err
	}

	working := filepath.Join(game, "Binaries", "Win32")
	err = os.MkdirAll(filepath.Join(game, LOGS_DIRECTORY), os.FileMode(0755))
	if err != nil {
		return fmt.Errorf("create the logs directory: %w", err)
	}

	// The game reads -abslog as a Windows path, so it has to be given one even though
	// everything around it here is a Unix path.
	arguments := gameArguments(windowsPath(gameLogPath(game)))
	command, err := runtimeCommand(game, log, filepath.Join(working, GAME_BINARY_NAME), arguments...)
	if err != nil {
		return err
	}
	command.Dir = working

	return command.Run()
}
