package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// CREATE_NO_WINDOW starts the cleanup process without flashing up a console window.
const CREATE_NO_WINDOW = 0x08000000

// scheduleCleanup launches a detached batch script that removes the install directory once
// this process exits and unlocks the last remaining file (this executable). The script
// retries until the removal succeeds, then deletes itself.
func scheduleCleanup(game string) {
	script := "@echo off\r\n" +
		":retry\r\n" +
		fmt.Sprintf("rmdir /s /q \"%s\" 2>nul\r\n", game) +
		fmt.Sprintf("if exist \"%s\" (\r\n", game) +
		"ping 127.0.0.1 -n 2 >nul\r\n" +
		"goto retry\r\n" +
		")\r\n" +
		"del /f /q \"%~f0\"\r\n"

	batch := filepath.Join(os.TempDir(), "tai-uninstall-cleanup.bat")
	if err := os.WriteFile(batch, []byte(script), os.FileMode(0644)); err != nil {
		return
	}

	cleanup := exec.Command("cmd", "/c", batch)
	// Run from the temp root so the process doesn't hold a handle on the directory it's
	// removing, and hide its console window.
	cleanup.Dir = os.TempDir()
	cleanup.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: CREATE_NO_WINDOW}
	_ = cleanup.Start()
}
