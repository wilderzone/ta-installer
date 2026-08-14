package main

import (
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
