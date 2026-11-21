package main

import (
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"unsafe"

	"github.com/jxeng/shortcut"
	"golang.org/x/sys/windows"
)

var procPropVariantClear = windows.NewLazySystemDLL("ole32.dll").NewProc("PropVariantClear")

// clearVariant frees whatever a property store put in a variant.
func clearVariant(variant *propVariant) {
	syscall.SyscallN(procPropVariantClear.Addr(), uintptr(unsafe.Pointer(variant)))
}

// readShortcutAppID reads a shortcut's AppUserModelID back out of its file, the way the play
// utility does. A shortcut without one reads as empty.
func readShortcutAppID(t *testing.T, path string) string {
	t.Helper()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	uninitialise, err := initialiseCOM()
	if err != nil {
		t.Fatalf("initialise COM: %v", err)
	}
	defer uninitialise()

	file, err := openShortcut(path, STGM_READ)
	if err != nil {
		t.Fatalf("open shortcut: %v", err)
	}
	defer file.release()
	store, err := file.queryInterface(&IID_PROPERTY_STORE)
	if err != nil {
		t.Fatalf("open the shortcut's properties: %v", err)
	}
	defer store.release()

	var variant propVariant
	hr, _, _ := syscall.SyscallN(store.methods[PROPERTY_STORE_GET_VALUE], uintptr(unsafe.Pointer(store)), uintptr(unsafe.Pointer(&PKEY_APP_USER_MODEL_ID)), uintptr(unsafe.Pointer(&variant)))
	if err := hresultError(hr); err != nil {
		t.Fatalf("read the AppUserModelID: %v", err)
	}
	defer clearVariant(&variant)
	if variant.kind != VT_LPWSTR || variant.value == nil {
		return ""
	}
	return windows.UTF16PtrToString(variant.value)
}

func TestGameShortcutsCarryAnAppIDEach(t *testing.T) {
	directory := t.TempDir()
	target, err := os.Executable()
	if err != nil {
		t.Fatalf("locate a target: %v", err)
	}
	game := shortcut.Shortcut{Target: target, WorkingDirectory: directory}
	pugs := game
	pugs.Arguments = "-hostx=" + LOGIN_SERVER_PUG

	// Twice, because verify writes the shortcuts again over the ones install wrote.
	for range 2 {
		if err := createGameShortcuts(directory, game, pugs); err != nil {
			t.Fatalf("create shortcuts: %v", err)
		}
		// Each login server's shortcut needs its own ID, or the taskbar couldn't tell which of
		// them to pin for a running game.
		for name, want := range map[string]string{SHORTCUT_NAME_MAIN: APP_ID_MAIN, SHORTCUT_NAME_PUG: APP_ID_PUG} {
			if got := readShortcutAppID(t, filepath.Join(directory, name)); got != want {
				t.Errorf("%s: got AppUserModelID %q, want %q", name, got, want)
			}
		}
	}
}

func TestShortcutsGoWithoutAnAppIDUnlessGivenOne(t *testing.T) {
	path := filepath.Join(t.TempDir(), UNINSTALL_SHORTCUT_NAME)
	target, err := os.Executable()
	if err != nil {
		t.Fatalf("locate a target: %v", err)
	}
	if err := createShortcut(shortcut.Shortcut{ShortcutPath: path, Target: target}, ""); err != nil {
		t.Fatalf("create shortcut: %v", err)
	}
	// The uninstaller isn't the game, so its windows mustn't be mistaken for the game's.
	if got := readShortcutAppID(t, path); got != "" {
		t.Errorf("got AppUserModelID %q, want none", got)
	}
}
