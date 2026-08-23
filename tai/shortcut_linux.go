package main

import (
	"errors"
	"fmt"
	"image/png"
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
const UNINSTALL_ICON_FILE = "uninstall.png"
const UNINSTALL_ICON_NAME = "tribes-ascend-uninstall"

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
	fmt.Fprintf(&file, "StartupWMClass=%s\n", escapeDesktopValue(GAME_BINARY_NAME))

	err := os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(file.String()), os.FileMode(0755))
}

// The sizes the hicolor theme indexes. A directory outside this set doesn't appear in the
// theme's index, and an icon filed there is simply never found.
var ICON_SIZES = [...]int{16, 22, 24, 32, 36, 48, 64, 72, 96, 128, 192, 256, 512}

// iconSizeDirectory returns the hicolor directory an icon of the given dimensions belongs in.
//
// Themes are indexed by directory, not by inspecting the images, so the directory is a claim
// about the icon's size. It snaps down to the nearest indexed size rather than up, because
// claiming an icon is larger than it is gets it chosen for contexts it can't fill, and shows
// up as a blurry upscale.
func iconSizeDirectory(width int, height int) string {
	// The larger edge decides: a non-square icon still has to fit the box it's given.
	size := max(width, height)
	chosen := ICON_SIZES[0]
	for _, candidate := range ICON_SIZES {
		if candidate <= size {
			chosen = candidate
		}
	}
	return fmt.Sprintf("%dx%d", chosen, chosen)
}

// removeInstalledIcons deletes the icon from every size directory it could be in. Which one
// was used depends on the dimensions of the icon that shipped, and those can change between
// releases, so a stale copy in another directory could otherwise outrank the current one.
func removeInstalledIcons(data string) {
	for _, size := range ICON_SIZES {
		_ = os.Remove(filepath.Join(data, "icons", "hicolor", fmt.Sprintf("%dx%d", size, size), "apps", ICON_NAME + ".png"))
		_ = os.Remove(filepath.Join(data, "icons", "hicolor", fmt.Sprintf("%dx%d", size, size), "apps", UNINSTALL_ICON_NAME + ".png"))
	}
}

// installIcons copies the icons shipped with the utilities into the user's icon theme and
// reports the names to reference them by. An installation whose utilities failed to download
// has no icons to install, so this reports empty names and the entries simply go without.
//
// The first returned string is the game icon, the second is the uninstall icon.
func installIcons(root string) (string, string) {
	iSource := filepath.Join(root, INSTALLATION_UTILITIES_DIRECTORY, ICON_FILE)
	uSource := filepath.Join(root, INSTALLATION_UTILITIES_DIRECTORY, UNINSTALL_ICON_FILE)
	iFile, err := os.Open(iSource)
	if err != nil {
		return "", ""
	}
	uFile, err := os.Open(uSource)
	if err != nil {
		return "", ""
	}
	// Only the header is read, so the size of the image doesn't matter here.
	iConfig, err := png.DecodeConfig(iFile)
	iFile.Close()
	if err != nil {
		return "", ""
	}
	uConfig, err := png.DecodeConfig(uFile)
	uFile.Close()
	if err != nil {
		return "", ""
	}

	data := dataDirectory()
	if data == "" {
		return "", ""
	}
	removeInstalledIcons(data)

	iTarget := filepath.Join(data, "icons", "hicolor", iconSizeDirectory(iConfig.Width, iConfig.Height), "apps", ICON_NAME + ".png")
	uTarget := filepath.Join(data, "icons", "hicolor", iconSizeDirectory(uConfig.Width, uConfig.Height), "apps", UNINSTALL_ICON_NAME + ".png")
	if err := os.MkdirAll(filepath.Dir(iTarget), os.FileMode(0755)); err != nil {
		return "", ""
	}
	if err := replaceInPlace(iSource, iTarget); err != nil {
		return "", ""
	}
	if err := os.MkdirAll(filepath.Dir(uTarget), os.FileMode(0755)); err != nil {
		return ICON_NAME, ""
	}
	if err := replaceInPlace(uSource, uTarget); err != nil {
		return ICON_NAME, ""
	}
	return ICON_NAME, UNINSTALL_ICON_NAME
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
// play utility rather than the game's executable directly, to account for the Linux
// compatibility layer.
func gameEntry(root string, icon string) desktopEntry {
	launcher := filepath.Join(root, INSTALLATION_UTILITIES_DIRECTORY, PLAY_NAME)
	return desktopEntry{
		Name:       PRODUCT_NAME,
		Comment:    SHORTCUT_DESCRIPTION,
		Exec:       quoteExecArgument(launcher),
		Icon:       icon,
		Path:       filepath.Join(root, "Binaries", "Win32"),
		Categories: "Game;ActionGame;",
		Terminal:   false,
	}
}

// createShortcuts creates desktop entries that launch the game. They're placed in the
// game's install directory, the applications menu, and on the user's desktop.
//
// They run the play utility rather than the game directly, so that a shortcut written by an
// older installation keeps working if the game's launch mechanism is ever changed.
func createShortcuts(logger *Logger, root string, includeOptionalShortcuts bool) error {
	if !fileExists(filepath.Join(root, INSTALLATION_UTILITIES_DIRECTORY, PLAY_NAME)) {
		return fmt.Errorf("create shortcuts: the play utility is missing")
	}

	icon, _ := installIcons(root)
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

	// The installed icon, from whichever size directory it landed in.
	if data := dataDirectory(); data != "" {
		removeInstalledIcons(data)
	}

	refreshDesktopDatabase()
	return firstErr
}
