package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// The icon shipped alongside the CLI, and the name it's installed under. Desktop
// environments look an unqualified Icon= name up in the icon theme, which is what lets the
// entry keep working if the installation directory is later moved.
const ICON_FILE = "icon.png"
const ICON_NAME = "tribes-ascend"

// desktopEntry is the subset of the Desktop Entry specification TAI writes.
// See: https://specifications.freedesktop.org/desktop-entry-spec/latest/
type desktopEntry struct {
	Name       string
	Comment    string
	Exec       string
	Icon       string
	Path       string
	Categories string
	Terminal   bool
}

// dataDirectory returns the XDG base directory for user-specific data files.
func dataDirectory() string {
	if data := os.Getenv("XDG_DATA_HOME"); data != "" {
		return data
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".local", "share")
}

// applicationsDirectory returns the directory desktop entries are read from for the
// applications menu.
func applicationsDirectory() string {
	data := dataDirectory()
	if data == "" {
		return ""
	}
	return filepath.Join(data, "applications")
}

// desktopDirectory returns the user's desktop. The location is configurable and localised,
// so xdg-user-dir is asked first; ~/Desktop is only a fallback for systems without
// xdg-user-dirs installed.
func desktopDirectory() string {
	output, err := exec.Command("xdg-user-dir", "DESKTOP").Output()
	if err == nil {
		path := strings.TrimSpace(string(output))
		// xdg-user-dir echoes $HOME back when the directory isn't configured, which is
		// its way of saying "there isn't one".
		if path != "" && path != os.Getenv("HOME") {
			return path
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, "Desktop")
}

// escapeDesktopValue escapes a string for use as a desktop entry value. Backslash is the
// escape character in the file format itself, so a literal one has to be doubled.
func escapeDesktopValue(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "\n", `\n`)
	value = strings.ReplaceAll(value, "\t", `\t`)
	value = strings.ReplaceAll(value, "\r", `\r`)
	return value
}

// quoteExecArgument quotes a single argument of an Exec value. Exec is a command line, not
// a path, so an argument containing a space (which the default install path does) has to be
// quoted or it would be read as two arguments. Within the quotes the spec reserves
// backslash, backtick, dollar and double-quote, each escaped with a backslash - and since
// those backslashes are themselves part of a desktop entry value, escapeDesktopValue
// doubles them again afterwards.
func quoteExecArgument(argument string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "`", "\\`", `$`, `\$`)
	return `"` + replacer.Replace(argument) + `"`
}

// writeDesktopEntry writes a desktop entry, replacing any existing one. The file is made
// executable because desktop environments refuse to launch an entry placed outside the
// applications directory - on the desktop, or in the installation directory - unless it is.
func writeDesktopEntry(path string, entry desktopEntry) error {
	var file strings.Builder
	file.WriteString("[Desktop Entry]\n")
	file.WriteString("Type=Application\n")
	file.WriteString("Version=1.0\n")
	fmt.Fprintf(&file, "Name=%s\n", escapeDesktopValue(entry.Name))
	if entry.Comment != "" {
		fmt.Fprintf(&file, "Comment=%s\n", escapeDesktopValue(entry.Comment))
	}
	fmt.Fprintf(&file, "Exec=%s\n", escapeDesktopValue(entry.Exec))
	if entry.Icon != "" {
		fmt.Fprintf(&file, "Icon=%s\n", escapeDesktopValue(entry.Icon))
	}
	if entry.Path != "" {
		fmt.Fprintf(&file, "Path=%s\n", escapeDesktopValue(entry.Path))
	}
	if entry.Categories != "" {
		fmt.Fprintf(&file, "Categories=%s\n", escapeDesktopValue(entry.Categories))
	}
	fmt.Fprintf(&file, "Terminal=%t\n", entry.Terminal)
	file.WriteString("StartupNotify=true\n")
	// Ties the game's window to this entry, so the taskbar shows its icon and name
	// rather than a generic Wine one.
	file.WriteString("StartupWMClass=TribesAscend.exe\n")

	err := os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(file.String()), os.FileMode(0755))
}

// installIcon copies the icon shipped with the utilities into the user's icon theme and
// reports the name to reference it by. An installation whose utilities failed to download
// has no icon to install, so this reports an empty name and the entries simply go without.
func installIcon(root string) string {
	source := filepath.Join(root, INSTALLATION_UTILITIES_DIRECTORY, ICON_FILE)
	if !fileExists(source) {
		return ""
	}
	data := dataDirectory()
	if data == "" {
		return ""
	}
	target := filepath.Join(data, "icons", "hicolor", "256x256", "apps", ICON_NAME + ".png")
	if err := os.MkdirAll(filepath.Dir(target), os.FileMode(0755)); err != nil {
		return ""
	}
	if err := replaceInPlace(source, target); err != nil {
		return ""
	}
	return ICON_NAME
}

// refreshDesktopDatabase asks the desktop environment to re-read the applications
// directory. Most environments notice the change by themselves, so this is best-effort and
// its absence is not an error.
func refreshDesktopDatabase() {
	applications := applicationsDirectory()
	if applications == "" {
		return
	}
	_ = exec.Command("update-desktop-database", applications).Run()
}

// gameEntry describes the desktop entry that launches the game. Launching goes through the
// CLI's "play" command rather than the executable directly, because on Linux the game needs
// a compatibility runtime configured around it - see runtime_linux.go.
func gameEntry(root string, icon string) desktopEntry {
	cli := filepath.Join(root, INSTALLATION_UTILITIES_DIRECTORY, CLI_NAME)
	return desktopEntry{
		Name:       PRODUCT_NAME,
		Comment:    SHORTCUT_DESCRIPTION,
		Exec:       fmt.Sprintf("%s play --path=%s", quoteExecArgument(cli), quoteExecArgument(root)),
		Icon:       icon,
		Path:       filepath.Join(root, "Binaries", "Win32"),
		Categories: "Game;ActionGame;",
		Terminal:   false,
	}
}

// createShortcuts creates desktop entries that launch the game. They're placed in the
// game's install directory, the applications menu, and on the user's desktop.
func createShortcuts(logger *Logger, root string, includeOptionalShortcuts bool) error {
	icon := installIcon(root)
	entry := gameEntry(root, icon)

	// Create an entry in the installation directory for convenience.
	err := writeDesktopEntry(filepath.Join(root, SHORTCUT_NAME), entry)
	if err != nil {
		return fmt.Errorf("create installation directory shortcut: %w", err)
	}

	if includeOptionalShortcuts {
		// Create an entry in the applications menu.
		applications := applicationsDirectory()
		if applications == "" {
			return fmt.Errorf("create applications menu shortcut: no home directory")
		}
		err = writeDesktopEntry(filepath.Join(applications, SHORTCUT_NAME), entry)
		if err != nil {
			return fmt.Errorf("create applications menu shortcut: %w", err)
		}

		// Create an entry on the desktop.
		desktop := desktopDirectory()
		if desktop == "" {
			return fmt.Errorf("create desktop shortcut: no home directory")
		}
		err = writeDesktopEntry(filepath.Join(desktop, SHORTCUT_NAME), entry)
		if err != nil {
			return fmt.Errorf("create desktop shortcut: %w", err)
		}

		refreshDesktopDatabase()
		logger.log(SHORTCUTS_SUCCESS)
	}

	return nil
}

// removeShortcuts deletes the desktop entries created by createShortcuts, along with the
// uninstall entry and the installed icon. It is best-effort: entries that are already gone
// are ignored, and it returns the first real error so the caller can warn without aborting
// the uninstall.
func removeShortcuts(root string) error {
	var firstErr error
	remove := func(path string) {
		if path == "" {
			return
		}
		err := os.Remove(path)
		if err != nil && !errors.Is(err, fs.ErrNotExist) && firstErr == nil {
			firstErr = err
		}
	}

	// The in-directory entries (may already be gone if the directory was cleared).
	remove(filepath.Join(root, SHORTCUT_NAME))
	remove(filepath.Join(root, UNINSTALL_SHORTCUT_NAME))

	// The applications menu entries.
	if applications := applicationsDirectory(); applications != "" {
		remove(filepath.Join(applications, SHORTCUT_NAME))
		remove(filepath.Join(applications, UNINSTALL_SHORTCUT_NAME))
	} else if firstErr == nil {
		firstErr = fmt.Errorf("locate the applications directory")
	}

	// The desktop entry.
	if desktop := desktopDirectory(); desktop != "" {
		remove(filepath.Join(desktop, SHORTCUT_NAME))
	} else if firstErr == nil {
		firstErr = fmt.Errorf("locate the desktop directory")
	}

	// The installed icon.
	if data := dataDirectory(); data != "" {
		remove(filepath.Join(data, "icons", "hicolor", "256x256", "apps", ICON_NAME + ".png"))
	}

	refreshDesktopDatabase()
	return firstErr
}
