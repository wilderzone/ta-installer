package main

import (
	"path/filepath"
)

func DownloadInstallationUtilities(logger *Logger, arguments arguments) error {
	// Check if the installation utilities already exist.
	if !installationUtilitiesAreMissing(arguments) {
		return nil
	}

	// Download.
	spinner := logger.spin("")
	err := Download(DEFAULT_REPO, arguments.path, INSTALLATION_UTILITIES)
	spinner.stop()
	if err != nil {
		return err
	}

	// TODO: Add the uninstall utility to the Registry so the game can be uninstalled through Windows.

	logger.log(UTILITIES_SUCCESS)
	return nil
}

func installationUtilitiesAreMissing(arguments arguments) bool {
	for _, file := range INSTALLATION_UTILITIES.Files {
		path := filepath.Join(arguments.path, INSTALLATION_UTILITIES_DIRECTORY, file.File)
		exists := fileExists(path)
		if !exists {
			return true
		}
	}

	return false
}
