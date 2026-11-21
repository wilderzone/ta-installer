package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// prepareRuntime does nothing here. Windows runs the game directly, so there is no
// compatibility layer to set up - the DirectX 9 runtime the game needs is installed by the
// installer, not by this.
func prepareRuntime(game string, log io.Writer) error {
	return nil
}

// launchGame starts the game and returns, rather than waiting for it to exit: nothing here
// needs to outlive the launch, and a launcher that lingers would keep a console window or a
// taskbar entry around for the whole session.
//
// The game manages its own log output; it isn't included in the launcher's logs.
func launchGame(game string, loginServer string, log io.Writer) error {
	working := filepath.Join(game, "Binaries", "Win32")
	path := filepath.Join(working, GAME_BINARY_NAME)
	arguments := append([]string{path}, gameArguments(loginServer)...)
	appID, pinnable := gameAppID(log)

	fmt.Fprintln(log, strings.Join(arguments, " "))

	process, err := startProcess(path, arguments, working, appID, pinnable)
	if err != nil {
		return err
	}
	return process.Release()
}

// startProcess starts a program with an AppUserModelID, which the taskbar uses to group windows.
func startProcess(path string, arguments []string, directory string, appID string, pinnable bool) (*os.Process, error) {
	application, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	commandLine, err := windows.UTF16PtrFromString(windows.ComposeCommandLine(arguments))
	if err != nil {
		return nil, err
	}
	working, err := windows.UTF16PtrFromString(directory)
	if err != nil {
		return nil, err
	}
	title, err := windows.UTF16PtrFromString(appID)
	if err != nil {
		return nil, err
	}

	startup := windows.StartupInfo{Title: title, Flags: STARTF_TITLEISAPPID}
	startup.Cb = uint32(unsafe.Sizeof(startup))
	if !pinnable {
		startup.Flags |= STARTF_PREVENTPINNING
	}
	var started windows.ProcessInformation
	err = windows.CreateProcess(application, commandLine, nil, nil, false, 0, nil, working, &startup, &started)
	if err != nil {
		return nil, fmt.Errorf("start %s: %w", path, err)
	}
	defer windows.CloseHandle(started.Thread)
	defer windows.CloseHandle(started.Process)

	// Found while the handle above still holds the process, so its ID can't have been reused.
	return os.FindProcess(int(started.ProcessId))
}
