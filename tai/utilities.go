package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"wilderzone.org/tai/internal/assets"
)

func downloadInstallationUtilities(ctx context.Context, logger *Logger, arguments arguments) error {
	// Check if the installation utilities already exist.
	if !installationUtilitiesAreMissing(arguments) {
		return nil
	}

	// Download.
	spinner := logger.spin("")
	err := download(ctx, DEFAULT_REPO, arguments.path, INSTALLATION_UTILITIES, nil)
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

// verifyUtilities checks the installed utilities (tai-cli.exe and friends) against the
// checksums published on the official CDN, and repairs them if the user agrees. It never
// returns an error: the game files have already been verified by the time it runs, so a
// missing network or a repair problem is reported and swallowed rather than failing verify.
func verifyUtilities(ctx context.Context, logger *Logger, arguments arguments) {
	// The checksums live on the CDN, not in this binary, so this needs the network. If it's
	// unavailable, the game files have still been verified - don't fail over it.
	util, err := getUtilitiesAsset(ctx, DEFAULT_REPO)
	if err != nil {
		// A cancelled run reports the interruption itself; this wasn't skipped.
		if ctx.Err() == nil {
			logger.log(UTILITIES_CHECK_SKIPPED, err.Error())
		}
		return
	}

	damaged := assets.VerifyAsset(ctx, arguments.path, util)
	if ctx.Err() != nil {
		return
	}
	if len(damaged) == 0 {
		logger.log(UTILITIES_VERIFIED)
		return
	}
	logger.log(UTILITIES_DAMAGED, assets.Describe(damaged))

	proceed := arguments.yes || logger.prompt(PROMPT_PROCEED)
	if !proceed {
		return
	}
	repairUtilities(ctx, logger, arguments.path, damaged)
}

// repairUtilities re-downloads the utilities package and replaces the damaged files. A
// running executable (tai-cli.exe, or verify.exe if it launched this run) can't be
// overwritten in place, so those are handed to a batch script that replaces them once this
// process exits. When TAI runs from elsewhere (such as the GUI's temp copy) nothing under
// the Installer directory is locked and every file is replaced directly.
func repairUtilities(ctx context.Context, logger *Logger, root string, damaged []FileResult) {
	// Download a fresh package into a staging directory. Nothing there is locked, so it
	// always unpacks cleanly to staging/Installer.
	staging, err := os.MkdirTemp("", "tai-utilities-*")
	if err != nil {
		logger.log(UTILITIES_FAILURE)
		logger.detail("stage utilities repair", err)
		return
	}
	err = download(ctx, DEFAULT_REPO, staging, INSTALLATION_UTILITIES, nil)
	if err != nil {
		_ = os.RemoveAll(staging)
		// A cancelled run reports the interruption itself; nothing has been touched yet.
		if ctx.Err() != nil {
			return
		}
		logger.log(UTILITIES_FAILURE)
		logger.detail("download utilities for repair", err)
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

// copyFile copies the file at sourcePath over the one at targetPath, creating it if it
// isn't there. It copies rather than renames so it works when the two are on different
// volumes, which they are whenever the source is a staging directory under the temp root.
func copyFile(sourcePath string, targetPath string, mode os.FileMode) error {
	in, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(targetPath, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, in)
	// Close the out file, regardless of copy errors.
	if closeErr := out.Close(); err == nil {
		err = closeErr
	}
	return err
}
