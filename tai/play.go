package main

import (
	"context"
)

// Handle the "play" command.
//
// The shortcuts an install creates point here rather than straight at the game, because on
// some platforms the game needs a compatibility layer configured around it. Keeping that in
// the CLI means a shortcut written by an older install keeps working after an update.
func Play(ctx context.Context, logger *Logger, arguments arguments) error {
	_, err := readManifest(logger, arguments.path)
	if err != nil {
		return userErrorf(err, "There's no installation at %s.", arguments.path)
	}

	return launchGame(ctx, logger, arguments.path)
}
