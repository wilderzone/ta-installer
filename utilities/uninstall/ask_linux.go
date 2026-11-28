package main

import "github.com/ncruces/zenity"

// ask asks whether to uninstall, and reports whether the user agreed. On Linux, the game keeps
// its settings inside the installation, so there's nothing more to ask about.
func ask(heading string, text string, icon zenity.DialogIcon) ([]string, bool) {
	return nil, askPlainly(heading, text, icon)
}
