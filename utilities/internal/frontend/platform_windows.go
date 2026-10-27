package frontend

import (
	"errors"
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

const CLI_NAME = "tai-cli.exe"

// CREATE_NO_WINDOW starts the CLI without a console window. The utilities have no console of
// their own, so Windows would otherwise open a fresh one for it.
const CREATE_NO_WINDOW = 0x08000000

// hideWindow keeps the CLI from opening a console window of its own.
func hideWindow(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: CREATE_NO_WINDOW}
}

// inTerminal is never needed here: zenity draws its own dialogs on Windows, so they're always available.
func inTerminal() bool {
	return false
}

// openTerminal is never needed here: zenity draws its own dialogs on Windows, so they're always available.
func openTerminal(command []string) error {
	return errors.New("no terminal to open")
}

// openFile opens a file in the user's preferred application for it.
func openFile(path string) error {
	verb, err := windows.UTF16PtrFromString("open")
	if err != nil {
		return err
	}
	file, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	return windows.ShellExecute(0, verb, file, nil, nil, windows.SW_SHOWNORMAL)
}
