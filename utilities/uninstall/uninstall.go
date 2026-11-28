package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ncruces/zenity"
	"wilderzone.org/tai-utilities/internal/frontend"
)

const TITLE = "Uninstall " + frontend.PRODUCT_NAME

func main() {
	os.Exit(frontend.Run(frontend.Utility{
		Command:      "uninstall",
		Title:        TITLE,
		Running:      "Uninstalling...",
		Success:      frontend.PRODUCT_NAME + " has been uninstalled.",
		Failure:      "Couldn't uninstall " + frontend.PRODUCT_NAME + ".",
		TemporaryLog: true,
		Confirm:      confirm,
		Complete:     removeLeftovers,
	}))
}

// confirm checks the installation before asking whether to remove it, so the question can
// be honest about what will be deleted. Cancel is the default answer. It returns the arguments
// the answer adds to the CLI's.
func confirm(found frontend.Installation) ([]string, bool) {
	status, err := found.Check()
	if err != nil {
		_ = zenity.Error(fmt.Sprintf("Couldn't check the installation at %s, so nothing was uninstalled.\n\n%s", found.Game, err.Error()), zenity.Title(TITLE))
		return nil, false
	}

	switch status {
		case frontend.CHECK_NOT_FOUND:
			_ = zenity.Info(fmt.Sprintf("%s isn't installed at %s.", frontend.PRODUCT_NAME, found.Game), zenity.Title(TITLE))
			return nil, false

		case frontend.CHECK_UNREADABLE:
			heading := fmt.Sprintf("%s's files couldn't be identified.", frontend.PRODUCT_NAME)
			text := fmt.Sprintf("Uninstalling anyway will permanently delete this folder and everything in it, including any files that weren't installed with the game:\n\n%s", found.Game)
			return ask(heading, text, zenity.WarningIcon)
	}

	heading := fmt.Sprintf("Are you sure you want to uninstall %s?", frontend.PRODUCT_NAME)
	text := fmt.Sprintf("This will permanently delete this folder and everything in it:\n\n%s", found.Game)
	return ask(heading, text, zenity.QuestionIcon)
}

// askPlainly asks whether to uninstall in a message box, and reports whether the user agreed.
func askPlainly(heading string, text string, icon zenity.DialogIcon) bool {
	err := zenity.Question(heading + "\n\n" + text,
		zenity.Title(TITLE),
		zenity.Icon(icon),
		zenity.OKLabel("Uninstall"),
		zenity.CancelLabel("Cancel"),
		zenity.DefaultCancel(),
	)
	return err == nil
}

// removeLeftovers removes what the CLI had to leave behind: the utilities directory this
// executable runs from, and the installation directory around it.
//
// It only does so when nothing else remains, so that whatever the CLI didn't remove - for
// whatever reason - is never swept up along with it.
func removeLeftovers(found frontend.Installation) {
	if onlyUtilitiesRemain(found) {
		scheduleCleanup(found.Game)
	}
}

// onlyUtilitiesRemain reports whether the utilities directory is the only thing left in the
// installation directory.
func onlyUtilitiesRemain(found frontend.Installation) bool {
	entries, err := os.ReadDir(found.Game)
	if err != nil {
		return false
	}
	utilities := filepath.Base(filepath.Dir(found.CLI))
	for _, entry := range entries {
		if entry.Name() != utilities {
			return false
		}
	}
	return true
}
