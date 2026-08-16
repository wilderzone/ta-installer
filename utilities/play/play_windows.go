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
// Because it returns straight away, the game's own output isn't this launcher's to collect;
// the log records what was started, and the game writes the rest itself through -abslog.
func launchGame(game string, log io.Writer) error {
	working := filepath.Join(game, "Binaries", "Win32")

	command := exec.Command(filepath.Join(working, GAME_BINARY_NAME), gameArguments(gameLogPath(game))...)
	command.Dir = working

	fmt.Fprintln(log, strings.Join(command.Args, " "))

	return command.Start()
}
