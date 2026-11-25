package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// removeSettings removes the folder the game keeps its settings and logs in. The game creates
// it in the user's Documents folder, outside the installation, so removing the installation
// leaves it behind. Every copy of the game on this user account shares it.
//
// Like removing the shortcuts, this is best-effort: a failure is reported, but doesn't fail the uninstall.
func removeSettings(logger *Logger) {
	directory, err := settingsDirectory()
	if err != nil {
		logger.log(SETTINGS_REMOVE_FAILURE, filepath.Join("Documents", SETTINGS_DIRECTORY))
		logger.detail("find the settings", err)
		return
	}
	err = os.RemoveAll(directory)
	if err != nil {
		logger.log(SETTINGS_REMOVE_FAILURE, directory)
		logger.detail("remove settings", err)
	}
}

// settingsDirectory returns the location of the game's settings. The Documents folder can be
// moved (eg, into OneDrive), so its location is retrieved from Windows rather than assumed.
func settingsDirectory() (string, error) {
	documents, err := getSpecialFolder("MyDocuments")
	if err != nil {
		return "", err
	}
	// Windows reports a folder it can't find as an empty path.
	if !filepath.IsAbs(documents) {
		return "", fmt.Errorf("the Documents folder's location %q isn't an absolute path", documents)
	}
	return filepath.Join(documents, SETTINGS_DIRECTORY), nil
}
