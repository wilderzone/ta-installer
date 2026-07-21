package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jxeng/shortcut"
)

// removeShortcuts deletes the shortcuts created by createShortcuts. It is
// best-effort: shortcuts that are already gone are ignored, and it returns the
// first real error so the caller can warn without aborting the uninstall.
func removeShortcuts(root string) error {
	var firstErr error
	remove := func(path string) {
		err := os.Remove(path)
		if err != nil && !errors.Is(err, fs.ErrNotExist) && firstErr == nil {
			firstErr = err
		}
	}

	// The in-directory shortcut (may already be gone if the directory was cleared).
	remove(filepath.Join(root, SHORTCUT_NAME))

	// The Desktop shortcut.
	if desktop, err := getSpecialFolder("Desktop"); err == nil {
		remove(filepath.Join(desktop, SHORTCUT_NAME))
	} else if firstErr == nil {
		firstErr = err
	}

	// The Start Menu folder (removes the folder and the shortcut inside it).
	if startMenu, err := getSpecialFolder("StartMenu"); err == nil {
		err = os.RemoveAll(filepath.Join(startMenu, "Programs", SHORTCUT_DIRECTORY))
		if err != nil && firstErr == nil {
			firstErr = err
		}
	} else if firstErr == nil {
		firstErr = err
	}

	return firstErr
}

// getSpecialFolder retrieves the real path for a given "Special Folder" in Windows.
// For a list of all Special Folders see: https://learn.microsoft.com/en-us/dotnet/api/system.environment.specialfolder
func getSpecialFolder(folder string) (string, error) {
	script := fmt.Sprintf("[Environment]::GetFolderPath([Environment+SpecialFolder]::%s)", folder)
	command := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	stdout, err := command.Output()
	if err != nil || len(stdout) == 0 {
		return "", fmt.Errorf("get Windows Special Folder %q: %w", folder, err)
	}
	path := strings.TrimSpace(string(stdout))
	return path, nil
}

// createShortcuts creates shortcuts to the game executable. The shortcuts are
// placed in the game's install directory, the Start Menu, and on the user's Desktop,
// and are preconfigured with the community server's host parameter.
func createShortcuts(logger *Logger, root string, includeOptionalShortcuts bool) error {
	cwd := filepath.Join(root, "Binaries", "Win32")
	target := filepath.Join(cwd, "TribesAscend.exe")
	icon := fmt.Sprintf("%s,0", target)
	logs := filepath.Join(root, LOGS_DIRECTORY, LOG_NAME)

	sc := shortcut.Shortcut{
		ShortcutPath:     "",
		Target:           target,
		IconLocation:     icon,
		Arguments:        fmt.Sprintf(SHORTCUT_ARGUMENTS, logs),
		Description:      SHORTCUT_DESCRIPTION,
		Hotkey:           "",
		WindowStyle:      "1",
		WorkingDirectory: cwd,
	}

	// Create a shortcut in the installation directory for convenience.
	sc.ShortcutPath = filepath.Join(root, SHORTCUT_NAME)
	err := shortcut.Create(sc)
	if err != nil {
		return fmt.Errorf("create installation directory shortcut: %w", err)
	}

	if includeOptionalShortcuts {
		logger.log(SHORTCUTS_CREATING)

		// Create a shortcut in the Start Menu.
		startMenuFolder, err := getSpecialFolder("StartMenu")
		if err != nil {
			return fmt.Errorf("create Start Menu shortcut: %w", err)
		}
		startMenuDirectory := filepath.Join(startMenuFolder, "Programs", SHORTCUT_DIRECTORY)
		// The Start Menu shortcut lives in its own folder, which Windows won't create for us.
		err = os.MkdirAll(startMenuDirectory, os.FileMode(0755))
		if err != nil {
			return fmt.Errorf("create Start Menu shortcut: %w", err)
		}
		sc.ShortcutPath = filepath.Join(startMenuDirectory, SHORTCUT_NAME)
		err = shortcut.Create(sc)
		if err != nil {
			return fmt.Errorf("create Start Menu shortcut: %w", err)
		}

		// Create a shortcut on the Desktop.
		desktopFolder, err := getSpecialFolder("Desktop")
		if err != nil {
			return fmt.Errorf("create Desktop shortcut: %w", err)
		}
		sc.ShortcutPath = filepath.Join(desktopFolder, SHORTCUT_NAME)
		err = shortcut.Create(sc)
		if err != nil {
			return fmt.Errorf("create Desktop shortcut: %w", err)
		}

		logger.log(SHORTCUTS_SUCCESS)
	}

	return nil
}
