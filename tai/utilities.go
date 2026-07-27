package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/jxeng/shortcut"
)

// CREATE_NO_WINDOW starts the deferred replace process without flashing up a console window.
const CREATE_NO_WINDOW = 0x08000000

func DownloadInstallationUtilities(logger *Logger, arguments arguments) error {
	// Check if the installation utilities already exist.
	if !installationUtilitiesAreMissing(arguments) {
		return nil
	}

	// Download.
	spinner := logger.spin("")
	err := Download(DEFAULT_REPO, arguments.path, INSTALLATION_UTILITIES)
	spinner.stop()
	if err != nil {
		return err
	}

	logger.log(UTILITIES_SUCCESS)
	return nil
}

func installationUtilitiesAreMissing(arguments arguments) bool {
	for _, file := range INSTALLATION_UTILITIES.Files {
		path := filepath.Join(arguments.path, INSTALLATION_UTILITIES_DIRECTORY, file.File)
		exists := fileExists(path)
		if !exists {
			return true
		}
	}

	return false
}

func createUninstallShortcut(root string) error {
	cwd := filepath.Join(root, INSTALLATION_UTILITIES_DIRECTORY)
	target := filepath.Join(cwd, "uninstall.exe")
	path := filepath.Join(root, "Uninstall.lnk")

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

	err := shortcut.Create(sc)
	if err != nil {
		return err
	}

	return nil
}

// verifyUtilities checks the installed utilities (tai-cli.exe and friends) against the
// checksums published on the official CDN, and repairs them if the user agrees. It never
// returns an error: the game files have already been verified by the time it runs, so a
// missing network or a repair problem is reported and swallowed rather than failing verify.
func verifyUtilities(logger *Logger, arguments arguments) {
	// The checksums live on the CDN, not in this binary, so this needs the network. If it's
	// unavailable, the game files have still been verified - don't fail over it.
	util, err := getUtilitiesAsset(DEFAULT_REPO)
	if err != nil {
		logger.log(UTILITIES_CHECK_SKIPPED, err.Error())
		return
	}

	damaged := verifyAsset(arguments.path, util)
	if len(damaged) == 0 {
		logger.log(UTILITIES_VERIFIED)
		return
	}
	logger.log(UTILITIES_DAMAGED, describe(damaged))

	proceed := arguments.yes || logger.prompt(PROMPT_PROCEED)
	if !proceed {
		return
	}
	repairUtilities(logger, arguments.path, damaged)
}

// repairUtilities re-downloads the utilities package and replaces the damaged files. A
// running executable (tai-cli.exe, or verify.exe if it launched this run) can't be
// overwritten in place, so those are handed to a batch script that replaces them once this
// process exits. When TAI runs from elsewhere (such as the GUI's temp copy) nothing under
// the Installer directory is locked and every file is replaced directly.
func repairUtilities(logger *Logger, root string, damaged []FileResult) {
	// Download a fresh package into a staging directory. Nothing there is locked, so it
	// always unpacks cleanly to staging/Installer.
	staging, err := os.MkdirTemp("", "tai-utilities-*")
	if err != nil {
		logger.log(UTILITIES_FAILURE)
		logger.detail("stage utilities repair", err)
		return
	}
	err = Download(DEFAULT_REPO, staging, INSTALLATION_UTILITIES)
	if err != nil {
		logger.log(UTILITIES_FAILURE)
		logger.detail("download utilities for repair", err)
		_ = os.RemoveAll(staging)
		return
	}

	source := filepath.Join(staging, INSTALLATION_UTILITIES_DIRECTORY)
	installer := filepath.Join(root, INSTALLATION_UTILITIES_DIRECTORY)
	// The directory may be gone entirely if every utility was missing.
	_ = os.MkdirAll(installer, os.FileMode(0777))
	repaired := []string{}
	locked := []string{}
	for _, result := range damaged {
		name := result.File.File
		if replaceInPlace(filepath.Join(source, name), filepath.Join(installer, name)) != nil {
			locked = append(locked, name)
		} else {
			repaired = append(repaired, name)
		}
	}

	if len(repaired) != 0 {
		logger.log(UTILITIES_REPAIRED, strings.Join(repaired, ", "))
		err := addUninstallerToRegistry(root)
		logger.detail("add uninstaller to registry", err)
	}

	if len(locked) != 0 {
		scheduleUtilitiesReplace(staging, installer, locked)
		logger.log(UTILITIES_REPAIR_DEFERRED, strings.Join(locked, ", "))
	} else {
		_ = os.RemoveAll(staging)
	}
}

// replaceInPlace overwrites the file at targetPath with the one at stagedPath. os.Create
// opens the target for writing, which fails with a sharing violation on a running
// executable - exactly how a locked utility is detected. It copies rather than renames, so
// it works when the staging directory is on a different volume from the installation.
func replaceInPlace(stagedPath string, targetPath string) error {
	in, err := os.Open(stagedPath)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	if closeErr := out.Close(); err == nil {
		err = closeErr
	}
	return err
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
