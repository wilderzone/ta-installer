package main

import (
	"bytes"
	"image"
	"image/png"
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
	path := filepath.Join(root, SHORTCUT_NAME)
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
	if err := os.WriteFile(filepath.Join(directory, ICON_FILE), encoded.Bytes(), os.FileMode(0644)); err != nil {
		t.Fatalf("write icon: %v", err)
	}
}

func TestInstallIconFilesByItsRealSize(t *testing.T) {
	data := t.TempDir()
	t.Setenv("XDG_DATA_HOME", data)
	root := t.TempDir()

	writeTestIcon(t, root, 48, 48)
	if name := installIcon(root); name != ICON_NAME {
		t.Fatalf("got %q, want %q", name, ICON_NAME)
	}
	installed := filepath.Join(data, "icons", "hicolor", "48x48", "apps", ICON_NAME + ".png")
	if _, err := os.Stat(installed); err != nil {
		t.Errorf("icon not installed at its own size: %v", err)
	}

	// A later release ships a different size. The old copy has to go, or it can outrank
	// the new one in the theme's lookup.
	writeTestIcon(t, root, 256, 256)
	if name := installIcon(root); name != ICON_NAME {
		t.Fatalf("got %q, want %q", name, ICON_NAME)
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
	if name := installIcon(t.TempDir()); name != "" {
		t.Errorf("got %q, want an empty name", name)
	}
}
