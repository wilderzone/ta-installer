package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestQuoteExecArgumentQuotesSpaces(t *testing.T) {
	// The default install path contains a space, so an unquoted Exec would be read as
	// two arguments and the game would never start.
	got := quoteExecArgument("/home/user/Games/Tribes Ascend")
	if got != `"/home/user/Games/Tribes Ascend"` {
		t.Errorf("got %s, want a quoted path", got)
	}
}

func TestQuoteExecArgumentEscapesReservedCharacters(t *testing.T) {
	got := quoteExecArgument(`/home/a"b$c` + "`d")
	// Each of " $ ` picks up a leading backslash, and the whole thing stays quoted.
	want := `"/home/a\"b\$c\` + "`" + `d"`
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestEscapeDesktopValueDoublesBackslashes(t *testing.T) {
	// Backslash is the escape character of the file format itself, so the backslashes
	// quoteExecArgument introduces have to survive being read back.
	if got := escapeDesktopValue(`a\b`); got != `a\\b` {
		t.Errorf("got %s, want a\\\\b", got)
	}
}

func TestWriteDesktopEntryIsLaunchable(t *testing.T) {
	root := filepath.Join(t.TempDir(), "Tribes Ascend")
	path := filepath.Join(root, SHORTCUT_NAME_MAIN)
	err := writeDesktopEntry(path, gameEntry(root, ICON_NAME))
	if err != nil {
		t.Fatalf("write desktop entry: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat desktop entry: %v", err)
	}
	// An entry outside the applications directory is ignored unless it's executable.
	if info.Mode().Perm() & 0111 == 0 {
		t.Errorf("got mode %v, want the entry to be executable", info.Mode().Perm())
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read desktop entry: %v", err)
	}
	entry := string(content)
	if !strings.HasPrefix(entry, "[Desktop Entry]\n") {
		t.Error("the entry must start with its group header")
	}
	// The install path has a space in it, so the launcher's path has to come back quoted.
	want := "Exec=" + quoteExecArgument(filepath.Join(root, INSTALLATION_UTILITIES_DIRECTORY, PLAY_NAME)) + "\n"
	if !strings.Contains(entry, want) {
		t.Errorf("got:\n%s\nwant an Exec line of:\n%s", entry, want)
	}
}

func TestIconSizeDirectorySnapsDown(t *testing.T) {
	cases := []struct {
		width  int
		height int
		want   string
	}{
		{256, 256, "256x256"}, // An indexed size is used as-is.
		{48, 48, "48x48"},
		{200, 200, "192x192"}, // Never rounded up: the theme would size it too large.
		{8, 8, "16x16"},       // Below the smallest indexed size, the smallest is the best on offer.
		{100, 60, "96x96"},    // The larger edge decides, since the icon has to fit its box.
		{1024, 1024, "512x512"},
	}
	for _, c := range cases {
		if got := iconSizeDirectory(c.width, c.height); got != c.want {
			t.Errorf("%dx%d: got %s, want %s", c.width, c.height, got, c.want)
		}
	}
}

// writeTestIcon writes a PNG of the given size where the utilities package would put it.
func writeTestIcon(t *testing.T, root string, width int, height int) {
	t.Helper()
	var encoded bytes.Buffer
	err := png.Encode(&encoded, image.NewRGBA(image.Rect(0, 0, width, height)))
	if err != nil {
		t.Fatalf("encode icon: %v", err)
	}
	directory := filepath.Join(root, INSTALLATION_UTILITIES_DIRECTORY)
	if err := os.MkdirAll(directory, os.FileMode(0755)); err != nil {
		t.Fatalf("create utilities directory: %v", err)
	}
	// Every icon ships in the utilities package, so a helper that wrote only some would be
	// testing a package that never gets published.
	for _, name := range []string{ICON_FILE, UNINSTALL_ICON_FILE, PUGS_ICON_FILE} {
		if err := os.WriteFile(filepath.Join(directory, name), encoded.Bytes(), os.FileMode(0644)); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
}

func TestInstallIconFilesByItsRealSize(t *testing.T) {
	data := t.TempDir()
	t.Setenv("XDG_DATA_HOME", data)
	root := t.TempDir()

	all := installedIcons{game: ICON_NAME, uninstall: UNINSTALL_ICON_NAME, pugs: PUGS_ICON_NAME}
	writeTestIcon(t, root, 48, 48)
	if icons := installIcons(root); icons != all {
		t.Fatalf("got %+v, want %+v", icons, all)
	}
	installed := filepath.Join(data, "icons", "hicolor", "48x48", "apps", ICON_NAME + ".png")
	if _, err := os.Stat(installed); err != nil {
		t.Errorf("icon not installed at its own size: %v", err)
	}

	// A later release ships a different size. The old copy has to go, or it can outrank
	// the new one in the theme's lookup.
	writeTestIcon(t, root, 256, 256)
	if icons := installIcons(root); icons != all {
		t.Fatalf("got %+v, want %+v", icons, all)
	}
	if _, err := os.Stat(filepath.Join(data, "icons", "hicolor", "256x256", "apps", ICON_NAME + ".png")); err != nil {
		t.Errorf("resized icon not installed: %v", err)
	}
	if _, err := os.Stat(installed); err == nil {
		t.Error("the previous size was left behind")
	}
}

func TestInstallIconWithoutOneReportsNoName(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	// An install whose utilities failed to download has no icon; entries go without.
	if icons := installIcons(t.TempDir()); icons != (installedIcons{}) {
		t.Errorf("got %+v, want empty names", icons)
	}
}

func TestPUGsEntryFallsBackToTheGameIcon(t *testing.T) {
	data := t.TempDir()
	t.Setenv("XDG_DATA_HOME", data)
	root := filepath.Join(t.TempDir(), "Tribes Ascend")
	installUtilities(t, root, PLAY_NAME)
	writeTestIcon(t, root, 32, 32)

	if err := createShortcuts(&Logger{silent: true}, root, false); err != nil {
		t.Fatalf("create shortcuts: %v", err)
	}
	entry, err := os.ReadFile(filepath.Join(root, SHORTCUT_NAME_PUG))
	if err != nil {
		t.Fatalf("read PUGs entry: %v", err)
	}
	if !strings.Contains(string(entry), "Icon=" + PUGS_ICON_NAME + "\n") {
		t.Errorf("got:\n%s\nwant the PUGs icon", entry)
	}

	// Utilities from before the PUGs icon existed: the entry borrows the game's icon rather
	// than going without, and the other icons are unaffected.
	if err := os.Remove(filepath.Join(root, INSTALLATION_UTILITIES_DIRECTORY, PUGS_ICON_FILE)); err != nil {
		t.Fatalf("remove PUGs icon: %v", err)
	}
	if err := createShortcuts(&Logger{silent: true}, root, false); err != nil {
		t.Fatalf("create shortcuts: %v", err)
	}
	entry, err = os.ReadFile(filepath.Join(root, SHORTCUT_NAME_PUG))
	if err != nil {
		t.Fatalf("read PUGs entry: %v", err)
	}
	if !strings.Contains(string(entry), "Icon=" + ICON_NAME + "\n") {
		t.Errorf("got:\n%s\nwant the game's icon", entry)
	}
	if _, err := os.Stat(filepath.Join(data, "icons", "hicolor", "32x32", "apps", ICON_NAME + ".png")); err != nil {
		t.Errorf("the game's icon wasn't installed: %v", err)
	}
}

// installUtilities creates empty stand-ins for the given installation utilities.
func installUtilities(t *testing.T, root string, names ...string) {
	t.Helper()
	directory := filepath.Join(root, INSTALLATION_UTILITIES_DIRECTORY)
	if err := os.MkdirAll(directory, os.FileMode(0755)); err != nil {
		t.Fatalf("create utilities directory: %v", err)
	}
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(directory, name), []byte{}, os.FileMode(0755)); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
}

// readEntry writes the entry and returns what was written.
func readEntry(t *testing.T, entry desktopEntry) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "entry.desktop")
	if err := writeDesktopEntry(path, entry); err != nil {
		t.Fatalf("write desktop entry: %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read desktop entry: %v", err)
	}
	return string(content)
}

func TestGameEntryOffersTheInstalledUtilities(t *testing.T) {
	root := filepath.Join(t.TempDir(), "Tribes Ascend")
	// An installation whose utilities predate update: it must not be offered.
	installUtilities(t, root, PLAY_NAME, VERIFY_NAME, UNINSTALL_NAME)

	entry := readEntry(t, gameEntry(root, ICON_NAME))
	if !strings.Contains(entry, "Actions=verify;uninstall;\n") {
		t.Errorf("got:\n%s\nwant actions for the installed utilities only", entry)
	}
	verify := quoteExecArgument(filepath.Join(root, INSTALLATION_UTILITIES_DIRECTORY, VERIFY_NAME))
	want := "\n[Desktop Action verify]\nName=Verify Files\nExec=" + escapeDesktopValue(verify) + "\n"
	if !strings.Contains(entry, want) {
		t.Errorf("got:\n%s\nwant a group for the verify action of:\n%s", entry, want)
	}
	if strings.Contains(entry, "[Desktop Action update]") {
		t.Error("got an action for a utility that isn't installed")
	}
	// The actions' groups have to come after the main one, or they'd swallow its keys.
	if strings.Index(entry, "StartupWMClass=") > strings.Index(entry, "[Desktop Action") {
		t.Errorf("got:\n%s\nwant every key of the main group before the actions", entry)
	}
}

func TestCreateShortcutsWritesAnEntryPerLoginServer(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	root := filepath.Join(t.TempDir(), "Tribes Ascend")
	installUtilities(t, root, PLAY_NAME)

	if err := createShortcuts(&Logger{silent: true}, root, false); err != nil {
		t.Fatalf("create shortcuts: %v", err)
	}

	launcher := quoteExecArgument(filepath.Join(root, INSTALLATION_UTILITIES_DIRECTORY, PLAY_NAME))
	cases := []struct {
		file string
		exec string
	}{
		// The main entry leaves the choice of server to the play utility.
		{SHORTCUT_NAME_MAIN, launcher},
		{SHORTCUT_NAME_PUG, launcher + " " + quoteExecArgument("-hostx=" + LOGIN_SERVER_PUG)},
	}
	for _, c := range cases {
		content, err := os.ReadFile(filepath.Join(root, c.file))
		if err != nil {
			t.Fatalf("read %s: %v", c.file, err)
		}
		// Entries sharing a file name would leave only the last one written.
		want := "Exec=" + escapeDesktopValue(c.exec) + "\n"
		if !strings.Contains(string(content), want) {
			t.Errorf("%s: got:\n%s\nwant an Exec line of:\n%s", c.file, content, want)
		}
	}
}

// isolateHome points everything outside the installation directory at a temporary home, so a
// test never touches the real desktop or applications menu.
func isolateHome(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(home, ".local", "share"))
}

// expectGameEntries checks whether a directory holds the game's entries, one for each login server.
func expectGameEntries(t *testing.T, directory string, want bool) {
	t.Helper()
	for _, name := range []string{SHORTCUT_NAME_MAIN, SHORTCUT_NAME_PUG} {
		path := filepath.Join(directory, name)
		_, err := os.Stat(path)
		if want && err != nil {
			t.Errorf("%s wasn't written: %v", path, err)
		}
		if !want && !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("got %s, want no entry there", path)
		}
	}
}

func TestShortcutsGoInTheMenuWhetherOrNotOnTheDesktop(t *testing.T) {
	for _, desktop := range []bool{true, false} {
		t.Run(fmt.Sprintf("desktop=%t", desktop), func(t *testing.T) {
			isolateHome(t)
			root := filepath.Join(t.TempDir(), "Tribes Ascend")
			installUtilities(t, root, PLAY_NAME)

			if err := createShortcuts(&Logger{silent: true}, root, desktop); err != nil {
				t.Fatalf("create shortcuts: %v", err)
			}
			// The choice only covers the desktop. The menu is where the game is looked for.
			expectGameEntries(t, applicationsDirectory(), true)
			expectGameEntries(t, desktopDirectory(), desktop)

			// And an uninstall takes them all away again.
			if err := removeShortcuts(root); err != nil {
				t.Fatalf("remove shortcuts: %v", err)
			}
			expectGameEntries(t, applicationsDirectory(), false)
			expectGameEntries(t, desktopDirectory(), false)
		})
	}
}

func TestOnlyTheGameEntryClaimsTheGameWindow(t *testing.T) {
	root := t.TempDir()
	if entry := readEntry(t, gameEntry(root, ICON_NAME)); !strings.Contains(entry, "StartupWMClass=" + GAME_BINARY_NAME + "\n") {
		t.Errorf("got:\n%s\nwant the game entry tied to the game's window", entry)
	}
	// A utility's window is a dialog, so an entry waiting for the game's window would leave
	// the desktop showing it as loading until it gave up.
	utility := VERIFY_UTILITY.entry(root, ICON_NAME)
	if entry := readEntry(t, utility); strings.Contains(entry, "StartupWMClass=") {
		t.Errorf("got:\n%s\nwant no window class for a utility's entry", entry)
	}
}

func TestUninstallEntryRunsTheUninstallUtility(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	root := filepath.Join(t.TempDir(), "Tribes Ascend")
	installUtilities(t, root, CLI_NAME, UNINSTALL_NAME)

	if err := createUninstallShortcut(root); err != nil {
		t.Fatalf("create uninstall shortcut: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(root, UNINSTALL_SHORTCUT_NAME))
	if err != nil {
		t.Fatalf("read uninstall entry: %v", err)
	}
	entry := string(content)
	uninstaller := quoteExecArgument(filepath.Join(root, INSTALLATION_UTILITIES_DIRECTORY, UNINSTALL_NAME))
	if !strings.Contains(entry, "Exec=" + escapeDesktopValue(uninstaller) + "\n") || !strings.Contains(entry, "Terminal=false\n") {
		t.Errorf("got:\n%s\nwant the uninstall utility, without a terminal", entry)
	}
}

func TestUninstallEntryFallsBackToTheCLI(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	root := filepath.Join(t.TempDir(), "Tribes Ascend")
	// Utilities from before the uninstall utility existed.
	installUtilities(t, root, CLI_NAME)

	if err := createUninstallShortcut(root); err != nil {
		t.Fatalf("create uninstall shortcut: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(root, UNINSTALL_SHORTCUT_NAME))
	if err != nil {
		t.Fatalf("read uninstall entry: %v", err)
	}
	entry := string(content)
	cli := quoteExecArgument(filepath.Join(root, INSTALLATION_UTILITIES_DIRECTORY, CLI_NAME))
	if !strings.Contains(entry, "Exec=" + escapeDesktopValue(cli) + " uninstall") || !strings.Contains(entry, "Terminal=true\n") {
		t.Errorf("got:\n%s\nwant the CLI, in a terminal", entry)
	}
}
