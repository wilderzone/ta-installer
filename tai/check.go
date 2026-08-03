package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
)

// Handle the "check" command.
func Check(logger *Logger, arguments arguments) error {
	_, err := readManifest(logger, arguments.path)
	if errors.Is(err, fs.ErrNotExist) {
		// No manifest means no installation. This is an expected outcome.
		logger.log(INSTALL_NOT_FOUND)
		return nil
	}
	if err != nil {
		// The manifest exists but couldn't be read - a genuine problem.
		return userErrorf(err, "Found an installation at %s, but its manifest couldn't be read.", arguments.path)
	}

	logger.log(INSTALL_FOUND)
	return nil
}

// Handle the "check-repo" command.
func CheckRepo(ctx context.Context, logger *Logger, arguments arguments) error {
	_, err := getManifest(ctx, logger, arguments.repo)
	if err != nil {
		// An abandoned request says nothing about whether the repository is there.
		if err := ctx.Err(); err != nil {
			return err
		}
		logger.log(REPO_NOT_FOUND)
		return nil
	}

	logger.log(REPO_FOUND)
	return nil
}

// Handle the "check-updates" command.
func CheckUpdates(ctx context.Context, logger *Logger, arguments arguments) error {
	logger.log(CHECKING_FOR_UPDATES)

	localManifest, err := readManifest(logger, arguments.path)
	if errors.Is(err, fs.ErrNotExist) {
		return userErrorf(err, "No installation found at %s. Run `install` first.", arguments.path)
	}
	if err != nil {
		return userErrorf(err, "The installation's manifest at %s couldn't be read.", arguments.path)
	}
	remoteManifest, err := getManifest(ctx, logger, arguments.repo)
	if err != nil {
		return userErrorf(err, "Couldn't check for updates from %s. Check your connection and that the repository URL is correct.", arguments.repo)
	}

	// Compare the version fields of the local and remote manifests.
	if localManifest.Version >= remoteManifest.Version {
		version := logger.colour(fmt.Sprintf("v%d", localManifest.Version), GREEN)
		logger.log(INSTALL_UP_TO_DATE, "Installation is up to date: " + version)
		return nil
	}
	old := logger.colour(fmt.Sprintf("v%d", localManifest.Version), RED)
	new := logger.colour(fmt.Sprintf("v%d", remoteManifest.Version), GREEN)
	logger.log(UPDATE_FOUND, fmt.Sprintf("%s → %s\n", old, new))

	return nil
}
