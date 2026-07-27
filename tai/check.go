package main

import (
	"errors"
	"io/fs"
)

// Handle the "check" command.
func Check(logger *Logger, arguments arguments) error {
	_, err := ReadManifest(logger, arguments.path)
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
func CheckRepo(logger *Logger, arguments arguments) error {
	_, err := GetManifest(logger, arguments.repo)
	if err != nil {
		logger.log(REPO_NOT_FOUND)
		return nil
	}

	logger.log(REPO_FOUND)
	return nil
}
