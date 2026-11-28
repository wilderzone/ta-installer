package main

import "github.com/ncruces/zenity"

// ask asks for confirmation when uninstalling.
//
// On Linux, the game's settings are inside the installation, so no additional settings-removal confirmation is necessary.
func ask(heading string, text string, icon zenity.DialogIcon) ([]string, bool) {
	return nil, askPlainly(heading, text, icon)
}
