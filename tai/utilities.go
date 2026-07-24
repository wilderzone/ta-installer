package main

import (
	"fmt"
	"path/filepath"

	"github.com/jxeng/shortcut"
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

	// Create a shortcut in the root directory for convenience.
	err = createUninstallShortcut(arguments.path)
	if err != nil {
		return fmt.Errorf("create uninstall shortcut: %w", err)
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

func createUninstallShortcut(root string) error {
	cwd := filepath.Join(root, INSTALLATION_UTILITIES_DIRECTORY)
	target := filepath.Join(cwd, "uninstall.exe")
	path := filepath.Join(root, "Uninstall.lnk")

	sc := shortcut.Shortcut{
		ShortcutPath:     path,
		Target:           target,
		IconLocation:     "%SystemRoot%\\System32\\SHELL32.dll,2",
		Arguments:        "",
		Description:      "Uninstall Tribes: Ascend",
		Hotkey:           "",
		WindowStyle:      "1",
		WorkingDirectory: cwd,
	}

	err := shortcut.Create(sc)
	if err != nil {
		return err
	}

	return nil
}
