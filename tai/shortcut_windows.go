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
//
// The shortcut is given an AppUserModelID if appID isn't empty. The library can't set one
// itself, so it's added to the file afterwards, on the same thread.
func createShortcut(sc shortcut.Shortcut, appID string) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	err := shortcut.Create(sc)
	if err != nil || appID == "" {
		return err
	}
	return setShortcutAppID(sc.ShortcutPath, appID)
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

	// The in-directory shortcuts (may already be gone if the directory was cleared).
	remove(filepath.Join(root, SHORTCUT_NAME_MAIN))
	remove(filepath.Join(root, SHORTCUT_NAME_PUG))

	// The Desktop shortcuts.
	if desktop, err := getSpecialFolder("Desktop"); err == nil {
		remove(filepath.Join(desktop, SHORTCUT_NAME_MAIN))
		remove(filepath.Join(desktop, SHORTCUT_NAME_PUG))
	} else if firstErr == nil {
		firstErr = err
	}

	// The Start Menu folder (removes the folder and the shortcuts inside it).
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

// bundledIcon returns the location of an icon shipped with the utilities, or a fallback if the icon doesn't exist.
func bundledIcon(root string, file string, fallback string) string {
	path := filepath.Join(root, INSTALLATION_UTILITIES_DIRECTORY, file)
	if !fileExists(path) {
		return fallback
	}
	return fmt.Sprintf("%s,0", path)
}

// createGameShortcuts creates the game's shortcuts in a directory: one for each login server.
//
// Each is given its own AppUserModelID, which the play utility passes on to the game. That
// way, the game's taskbar button is the shortcut's, and pinning the game pins the shortcut
// rather than the game's executable, which can't reach a login server by itself.
func createGameShortcuts(directory string, game shortcut.Shortcut, pugs shortcut.Shortcut) error {
	game.ShortcutPath = filepath.Join(directory, SHORTCUT_NAME_MAIN)
	err := createShortcut(game, APP_ID_MAIN)
	if err != nil {
		return err
	}
	pugs.ShortcutPath = filepath.Join(directory, SHORTCUT_NAME_PUG)
	return createShortcut(pugs, APP_ID_PUG)
}

// createShortcuts creates shortcuts that launch the game, one for each login server. The
// shortcuts are placed in the game's install directory and the Start Menu, and on the user's
// Desktop if preferred.
//
// They run the play utility rather than the game directly, so that a shortcut written by an
// older installation keeps working if the game's launch mechanism is ever changed.
func createShortcuts(logger *Logger, root string, includeDesktopShortcuts bool) error {
	cwd := filepath.Join(root, "Binaries", "Win32")
	gameIcon := fmt.Sprintf("%s,0", filepath.Join(cwd, GAME_BINARY_NAME))
	icon := bundledIcon(root, ICON_FILE, gameIcon)
	launcher := filepath.Join(root, INSTALLATION_UTILITIES_DIRECTORY, PLAY_NAME)

	if !fileExists(launcher) {
		return fmt.Errorf("create shortcuts: the play utility is missing")
	}

	sc := shortcut.Shortcut{
		ShortcutPath:     "",
		Target:           launcher,
		IconLocation:     icon,
		Arguments:        "",
		Description:      SHORTCUT_DESCRIPTION,
		Hotkey:           "",
		WindowStyle:      "7", // Minimised, in case the launcher was built with a console.
		WorkingDirectory: cwd,
	}

	// An alternate shortcut for the PUG login server.
	pugs := sc
	pugs.Arguments = fmt.Sprintf("-hostx=%s", LOGIN_SERVER_PUG)
	pugs.IconLocation = bundledIcon(root, PUGS_ICON_FILE, gameIcon)

	// Create shortcuts in the installation directory for convenience.
	err := createGameShortcuts(root, sc, pugs)
	if err != nil {
		return fmt.Errorf("create installation directory shortcut: %w", err)
	}

	// Create shortcuts in the Start Menu. They're created regardless of user preference,
	// since the Start Menu is where Windows expects to find links to an installed application (for taskbar pinning, etc).
	startMenuFolder, err := getSpecialFolder("StartMenu")
	if err != nil {
		return fmt.Errorf("create Start Menu shortcut: %w", err)
	}
	startMenuDirectory := filepath.Join(startMenuFolder, "Programs", SHORTCUT_DIRECTORY)
	// The Start Menu shortcuts live in their own folder, which Windows won't create for us.
	err = os.MkdirAll(startMenuDirectory, os.FileMode(0755))
	if err != nil {
		return fmt.Errorf("create Start Menu shortcut: %w", err)
	}
	err = createGameShortcuts(startMenuDirectory, sc, pugs)
	if err != nil {
		return fmt.Errorf("create Start Menu shortcut: %w", err)
	}

	if includeDesktopShortcuts {
		// Create shortcuts on the Desktop.
		desktopFolder, err := getSpecialFolder("Desktop")
		if err != nil {
			return fmt.Errorf("create Desktop shortcut: %w", err)
		}
		err = createGameShortcuts(desktopFolder, sc, pugs)
		if err != nil {
			return fmt.Errorf("create Desktop shortcut: %w", err)
		}
	}

	logger.log(SHORTCUTS_SUCCESS)
	return nil
}
