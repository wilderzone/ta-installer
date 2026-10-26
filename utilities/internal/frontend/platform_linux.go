package frontend

import "os/exec"

const CLI_NAME = "tai-cli"

// hideWindow does nothing here. A program started without a terminal doesn't get one.
func hideWindow(command *exec.Cmd) {}

// openFile opens a file in the user's preferred application for it.
func openFile(path string) error {
	return exec.Command("xdg-open", path).Start()
}
