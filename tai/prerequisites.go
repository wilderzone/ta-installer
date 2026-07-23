package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func Prerequisites(logger *Logger, arguments arguments) error {
	// Check if d3dx9 is already installed. Note that the Windows directory must be
	// joined from a rooted path: filepath.Join("C:", ...) yields the drive-relative
	// "C:Windows", which never resolves.
	windowsDirectory := os.Getenv("SystemRoot")
	if windowsDirectory == "" {
		windowsDirectory = filepath.Join("C:\\", "Windows")
	}
	d3dx9Path := filepath.Join(windowsDirectory, "SysWOW64", "d3dx9_24.dll")
	d3dExists := fileExists(d3dx9Path)
	if d3dExists {
		return nil
	}

	logger.log(PREREQUISITES_REQUIRED)

	// Download the prerequisites asset.
	logger.log(PREREQUISITES_DOWNLOADING)
	err := Download(arguments.repo, arguments.path, PREREQUISITES)
	if err != nil {
		return userErrorf(err, "Couldn't install the DirectX 9 package.")
	}

	logger.log(PREREQUISITES_INSTALLING)

	// Run both DX9 installation stages. The first stage only unpacks the
	// redistributable into the temp directory, so it needs no special permissions.
	prerequisite1 := filepath.Join(arguments.path, "Binaries", "Redist", "directx_Jun2010_redist.exe")
	command1 := exec.Command(prerequisite1, "/T:" + filepath.Join(arguments.path, TEMP_DIRECTORY, "dx"), "/Q")
	err = command1.Run()
	if err != nil {
		return fmt.Errorf("extract DirectX 9 redistributable: %w", err)
	}

	// The second stage writes into the Windows directory, so it must be elevated.
	// This is the only step that requires administrator permission, so the user is
	// only prompted when DirectX 9 is genuinely missing.
	//
	// Its exit code is treated as a hint rather than a verdict: the installed
	// runtime is the real signal, so a non-zero exit is not fatal until the check
	// below has also failed. This tolerates codes that report success in disguise,
	// such as 3010 (success, reboot required).
	prerequisite2 := filepath.Join(arguments.path, TEMP_DIRECTORY, "dx", "DXSETUP.exe")
	setupErr := runElevated(prerequisite2, "/silent")

	// Confirm the runtime the game actually needs is present.
	d3dExists = fileExists(d3dx9Path)
	if d3dExists {
		return nil
	}

	if setupErr != nil {
		return fmt.Errorf("run DirectX 9 setup: %w", setupErr)
	}
	return userErrorf(nil, "DirectX 9 setup finished, but the DirectX 9 runtime still isn't installed. Try installing DirectX 9 manually, then run the installer again.")
}
