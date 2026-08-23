package main

import (
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"strings"
)

// prepareRuntime does nothing here. Windows runs the game directly, so there is no
// compatibility layer to set up - the DirectX 9 runtime the game needs is installed by the
// installer, not by this.
func prepareRuntime(game string, log io.Writer) error {
	return nil
}

// launchGame starts the game and returns, rather than waiting for it to exit: nothing here
// needs to outlive the launch, and a launcher that lingers would keep a console window or a
// taskbar entry around for the whole session.
//
// The game manages its own log output; it isn't included in the launcher's logs.
func launchGame(game string, log io.Writer) error {
	working := filepath.Join(game, "Binaries", "Win32")

	command := exec.Command(filepath.Join(working, GAME_BINARY_NAME), gameArguments(gameLogPath(game))...)
	command.Dir = working

	fmt.Fprintln(log, strings.Join(command.Args, " "))

	return command.Start()
}
