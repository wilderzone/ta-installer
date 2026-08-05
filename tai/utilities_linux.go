package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// replaceInPlace overwrites the file at targetPath with the one at stagedPath.
//
// It writes a sibling and renames over the target rather than truncating it. Truncating
// would fail with ETXTBSY on a running executable - which the CLI repairing itself always
// is - whereas a rename only swaps the directory entry and leaves the running inode alone.
// So on Linux this always succeeds, nothing is ever left locked, and the deferred-replace
// path below is unreachable.
func replaceInPlace(stagedPath string, targetPath string) error {
	// The temporary lands beside the target so the rename stays within one filesystem.
	// Renaming across filesystems fails with EXDEV.
	temporary := targetPath + ".tai-new"
	err := copyFile(stagedPath, temporary, os.FileMode(0755))
	if err != nil {
		return err
	}
	err = os.Rename(temporary, targetPath)
	if err != nil {
		_ = os.Remove(temporary)
		return err
	}
	return nil
}

// scheduleUtilitiesReplace exists only to satisfy the shared repair path. Nothing on Linux
// can be locked open (see replaceInPlace), so this is never reached; it cleans up the
// staging directory rather than silently leaving it behind if it somehow is.
func scheduleUtilitiesReplace(staging string, installer string, files []string) {
	_ = os.RemoveAll(staging)
}

// createUninstallShortcut writes the desktop entry that uninstalls the game. Windows lists
// an installation in Add/Remove Programs; on Linux the applications menu is the equivalent
// place to look, so the entry goes there as well as in the installation directory.
func createUninstallShortcut(root string) error {
	cli := filepath.Join(root, INSTALLATION_UTILITIES_DIRECTORY, CLI_NAME)
	entry := desktopEntry{
		Name:       "Uninstall " + PRODUCT_NAME,
		Comment:    "Remove " + PRODUCT_NAME + " and its compatibility runtime",
		Exec:       fmt.Sprintf("%s uninstall --path=%s", quoteExecArgument(cli), quoteExecArgument(root)),
		Icon:       ICON_NAME,
		Categories: "Game;",
		// Uninstalling asks for confirmation and reports what it removed, so it needs
		// somewhere to ask and somewhere to print.
		Terminal:   true,
	}

	err := writeDesktopEntry(filepath.Join(root, UNINSTALL_SHORTCUT_NAME), entry)
	if err != nil {
		return err
	}

	applications := applicationsDirectory()
	if applications == "" {
		return fmt.Errorf("locate the applications directory")
	}
	err = writeDesktopEntry(filepath.Join(applications, UNINSTALL_SHORTCUT_NAME), entry)
	if err != nil {
		return err
	}

	refreshDesktopDatabase()
	return nil
}
