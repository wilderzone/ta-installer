package frontend

import (
	"errors"
	"os"
	"os/exec"

	"golang.org/x/sys/unix"
)

const CLI_NAME = "tai-cli"

// TERMINALS are the terminal emulators a utility can reopen itself in, in order of preference,
// each with the arguments that come before the command it should run.
//
// The user's own choice comes first: xdg-terminal-exec where it's installed, and Debian's
// x-terminal-emulator alternative. Then the defaults of the common desktops, then the rest.
var TERMINALS = []struct {
	name      string
	arguments []string
}{
	{"xdg-terminal-exec", []string{"--"}},
	{"x-terminal-emulator", []string{"-e"}},
	{"ptyxis", []string{"--"}},
	{"kgx", []string{"--"}},
	{"gnome-terminal", []string{"--"}},
	{"konsole", []string{"-e"}},
	{"xfce4-terminal", []string{"-x"}},
	{"mate-terminal", []string{"-x"}},
	{"alacritty", []string{"-e"}},
	{"kitty", []string{}},
	{"foot", []string{}},
	{"wezterm", []string{"start", "--"}},
	{"xterm", []string{"-e"}},
}

// inTerminal reports whether the utility is running in an interactive terminal, where the
// user can read the CLI's output and answer its questions.
//
// Being a character device isn't enough: a desktop can start a program with /dev/null as
// its input and output, which is one too. Only a terminal answers a terminal's ioctl.
func inTerminal() bool {
	for _, stream := range []*os.File{os.Stdin, os.Stdout} {
		_, err := unix.IoctlGetTermios(int(stream.Fd()), unix.TCGETS)
		if err != nil {
			return false
		}
	}
	return true
}

// hideWindow does nothing here. A program started without a terminal doesn't get one.
func hideWindow(command *exec.Cmd) {}

// openFile opens a file in the user's preferred application for it.
func openFile(path string) error {
	return exec.Command("xdg-open", path).Start()
}

// openTerminal runs a command in a new window of the first terminal emulator it finds. It
// doesn't wait for the command to finish: some terminals hand it to an existing process and
// return straight away, so there's nothing to wait for.
func openTerminal(command []string) error {
	for _, terminal := range TERMINALS {
		path, err := exec.LookPath(terminal.name)
		if err != nil {
			continue
		}
		arguments := append(append([]string{}, terminal.arguments...), command...)
		return exec.Command(path, arguments...).Start()
	}
	return errors.New("there's no way to show dialogs on this system (install zenity), and no terminal to run in instead")
}
