package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/jxeng/shortcut"
)

// CREATE_NO_WINDOW starts the deferred replace process without flashing up a console window.
const CREATE_NO_WINDOW = 0x08000000

// replaceInPlace overwrites the file at targetPath with the one at stagedPath. Opening the
// target for writing fails with a sharing violation on a running executable - exactly how a
// locked utility is detected, so the caller can defer those to scheduleUtilitiesReplace.
func replaceInPlace(stagedPath string, targetPath string) error {
	return copyFile(stagedPath, targetPath, os.FileMode(0666))
}

func createUninstallShortcut(root string) error {
	cwd := filepath.Join(root, INSTALLATION_UTILITIES_DIRECTORY)
	target := filepath.Join(cwd, "uninstall.exe")
	path := filepath.Join(root, UNINSTALL_SHORTCUT_NAME)

	sc := shortcut.Shortcut{
		ShortcutPath:     path,
		Target:           target,
		IconLocation:     "%SystemRoot%\\System32\\SHELL32.dll,2",
		Arguments:        "",
		Description:      "Uninstall Tribes: Ascend",
		Hotkey:           "",
		WindowStyle:      "1",
		WorkingDirectory: cwd,
	}

	err := createShortcut(sc)
	if err != nil {
		return err
	}

	return nil
}

// scheduleUtilitiesReplace launches a detached batch script that copies the freshly staged
// files over the locked ones once this process exits and releases them, then removes the
// staging directory and deletes itself. It mirrors the uninstaller's deferred cleanup.
func scheduleUtilitiesReplace(staging string, installer string, files []string) {
	source := filepath.Join(staging, INSTALLATION_UTILITIES_DIRECTORY)

	var script strings.Builder
	script.WriteString("@echo off\r\n:retry\r\nset FAILED=\r\n")
	for _, name := range files {
		fmt.Fprintf(&script, "copy /y \"%s\" \"%s\" >nul 2>&1 || set FAILED=1\r\n",
			filepath.Join(source, name), filepath.Join(installer, name))
	}
	script.WriteString("if defined FAILED (\r\nping 127.0.0.1 -n 2 >nul\r\ngoto retry\r\n)\r\n")
	fmt.Fprintf(&script, "rmdir /s /q \"%s\" >nul 2>&1\r\n", staging)
	script.WriteString("del /f /q \"%~f0\"\r\n")

	batch := filepath.Join(os.TempDir(), "tai-utilities-repair.bat")
	if err := os.WriteFile(batch, []byte(script.String()), os.FileMode(0644)); err != nil {
		return
	}

	replace := exec.Command("cmd", "/c", batch)
	// Run from the temp root so the process doesn't hold a handle on the staging directory it
	// removes, and hide its console window.
	replace.Dir = os.TempDir()
	replace.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: CREATE_NO_WINDOW}
	_ = replace.Start()
}
