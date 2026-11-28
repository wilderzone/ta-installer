package main

import (
	"os"
	"path/filepath"

	"github.com/ncruces/zenity"
	"golang.org/x/sys/windows"
)

// SETTINGS_DIRECTORY is where the game keeps its settings and logs, relative to the user's Documents folder.
const SETTINGS_DIRECTORY = `My Games\Tribes Ascend`
const SETTINGS_CHECKBOX_MESSAGE = "Also remove the game's settings and logs from:\nDocuments\\" + SETTINGS_DIRECTORY

// ask asks whether to uninstall, and reports whether the user agreed, along with any arguments
// their answer adds to the CLI's.
//
// On Windows, the game keeps its settings outside the installation. If they exist, the
// question comes with a checkbox to delete them too, which starts unticked.
func ask(heading string, text string, icon zenity.DialogIcon) ([]string, bool) {
	dialog := taskDialog{
		title:   TITLE,
		heading: heading,
		text:    text,
		accept:  "Uninstall",
		warning: icon == zenity.WarningIcon,
	}
	if settingsExist() {
		dialog.checkbox = SETTINGS_CHECKBOX_MESSAGE
	}

	accepted, checked, err := dialog.show()
	if err != nil {
		return nil, askPlainly(heading, text, icon)
	}
	if accepted && checked {
		return []string{"--remove-settings"}, true
	}
	return nil, accepted
}

// settingsExist reports whether the game has left any settings in the user's Documents folder.
// Documents can be moved (eg, into OneDrive), so its location is retrieved from Windows.
func settingsExist() bool {
	documents, err := windows.KnownFolderPath(windows.FOLDERID_Documents, windows.KF_FLAG_DEFAULT)
	if err != nil {
		return false
	}
	_, err = os.Stat(filepath.Join(documents, SETTINGS_DIRECTORY))
	return err == nil
}
