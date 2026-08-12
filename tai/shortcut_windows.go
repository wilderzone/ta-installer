package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jxeng/shortcut"
)

// createShortcut creates a single shortcut, pinning the goroutine to its OS thread for the
// duration. github.com/jxeng/shortcut initialises COM (CoInitializeEx) on the current
// thread and then makes a series of COM calls, but COM state is per-thread and Go's
// scheduler may move a goroutine to a different OS thread between calls - especially after
// the I/O that precedes shortcut creation during an install. When that happens the COM
// calls land on a thread where COM was never initialised, giving the intermittent
// "CoInitialize has not been called" error. Locking the thread keeps the whole sequence
// on one thread.
func createShortcut(sc shortcut.Shortcut) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	return shortcut.Create(sc)
}

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

// createShortcuts creates shortcuts that launch the game. The shortcuts are placed in the
// game's install directory, the Start Menu, and on the user's Desktop.
//
// They run the CLI's "play" command rather than the game directly, so that a shortcut
// written by an older installation keeps working when the way the game is launched changes -
// the launch arguments live in one place, and an update fixes them for every existing
// shortcut at once. The icon still comes from the game executable, so nothing looks different.
func createShortcuts(logger *Logger, root string, includeOptionalShortcuts bool) error {
	cwd := filepath.Join(root, "Binaries", "Win32")
	icon := fmt.Sprintf("%s,0", filepath.Join(cwd, GAME_BINARY_NAME))
	cli := filepath.Join(root, INSTALLATION_UTILITIES_DIRECTORY, CLI_NAME)

	// Downloading the utilities is best-effort, so the shortcuts' target may not be there.
	// The caller reports this as a shortcut failure rather than aborting the install.
	if !fileExists(cli) {
		return fmt.Errorf("the installation utilities are missing, so a shortcut would have nothing to launch")
	}

	sc := shortcut.Shortcut{
		ShortcutPath:     "",
		Target:           cli,
		IconLocation:     icon,
		// %q quotes the path, which the default installation directory needs: this is a
		// command line, so an unquoted space would split it into two arguments.
		Arguments:        fmt.Sprintf("play --path=%q", root),
		Description:      SHORTCUT_DESCRIPTION,
		Hotkey:           "",
		WindowStyle:      "7", // Minimised, so the CLI's console doesn't take focus.
		WorkingDirectory: cwd,
	}

	// Create a shortcut in the installation directory for convenience.
	sc.ShortcutPath = filepath.Join(root, SHORTCUT_NAME)
	err := createShortcut(sc)
	if err != nil {
		return fmt.Errorf("create installation directory shortcut: %w", err)
	}

	if includeOptionalShortcuts {
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
		err = createShortcut(sc)
		if err != nil {
			return fmt.Errorf("create Start Menu shortcut: %w", err)
		}

		// Create a shortcut on the Desktop.
		desktopFolder, err := getSpecialFolder("Desktop")
		if err != nil {
			return fmt.Errorf("create Desktop shortcut: %w", err)
		}
		sc.ShortcutPath = filepath.Join(desktopFolder, SHORTCUT_NAME)
		err = createShortcut(sc)
		if err != nil {
			return fmt.Errorf("create Desktop shortcut: %w", err)
		}

		logger.log(SHORTCUTS_SUCCESS)
	}

	return nil
}
