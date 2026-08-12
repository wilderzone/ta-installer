package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// launchGame starts the game. Windows runs it directly - there's no compatibility layer to
// go through - so this is the same command the shortcut used to hold, kept here so both
// platforms launch the same way.
func launchGame(ctx context.Context, logger *Logger, root string) error {
	working := filepath.Join(root, "Binaries", "Win32")
	logs := filepath.Join(root, LOGS_DIRECTORY, LOG_NAME)
	err := os.MkdirAll(filepath.Dir(logs), os.FileMode(0755))
	if err != nil {
		return fmt.Errorf("create the logs directory: %w", err)
	}

	arguments := strings.Fields(fmt.Sprintf(SHORTCUT_ARGUMENTS, logs))
	command := exec.CommandContext(ctx, filepath.Join(working, GAME_BINARY_NAME), arguments...)
	command.Dir = working

	logger.log(LAUNCHING)
	return command.Start()
}
