package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true // File exists.
	}
	if errors.Is(err, os.ErrNotExist) {
		return false // File explicitly does not exist.
	}
	// File may or may not exist (e.g., permission denied, disk failure).
	return false 
}

func Download(repo string, root string, asset Asset) error {
	// Download the asset into the temp directory.
	tempDirectory := filepath.Join(root, TEMP_DIRECTORY)
	err := os.MkdirAll(tempDirectory, os.FileMode(0777))
	if err != nil {
		return fmt.Errorf("create temp directory: %w", err)
	}
	file, err := os.CreateTemp(tempDirectory, "download-*")
	fileName := file.Name()
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	source, err := url.JoinPath(repo, asset.Source)
	if err != nil {
		return fmt.Errorf("build source URL: %w", err)
	}
	response, err := http.Get(source)
	if err != nil {
		return fmt.Errorf("request %s: %w", source, err)
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("request %s: unexpected status %s", source, response.Status)
	}
	defer response.Body.Close()
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return fmt.Errorf("download %s: %w", source, err)
	}
	err = file.Close()
	if err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	// Unpack the downloaded file into the destination directory.
	err = Unpack(root, fileName, asset)
	if err != nil {
		return fmt.Errorf("unpack %s: %w", asset.Name, err)
	}
	err = os.Remove(fileName)
	if err != nil {
		return fmt.Errorf("remove temp file: %w", err)
	}

	return nil
}

func Unpack(root string, archivePath string, asset Asset) error {
	destination := filepath.Join(root, asset.Destination)
	err := os.MkdirAll(destination, os.FileMode(0777))
	if err != nil {
		return fmt.Errorf("create destination directory: %w", err)
	}
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		// Prevent zip-slips.
		path := filepath.Join(destination, file.Name)
		isWithinDestination := strings.HasPrefix(path, filepath.Clean(destination) + string(os.PathSeparator))
		if !isWithinDestination {
			return fmt.Errorf("illegal path in archive: %q", file.Name)
		}

		// Handle directories.
		if file.FileInfo().IsDir() {
			err = os.MkdirAll(path, os.FileMode(0777))
			if err != nil {
				return fmt.Errorf("create directory %s: %w", file.Name, err)
			}
			continue
		}

		// Open the destination file.
		output, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("create %s: %w", file.Name, err)
		}

		// Open the archived file.
		archived, err := file.Open()
		if err != nil {
			output.Close()
			return fmt.Errorf("read %s from archive: %w", file.Name, err)
		}

		// Copy the file content.
		_, err = io.Copy(output, archived)
		output.Close()
		archived.Close()

		if err != nil {
			return fmt.Errorf("write %s: %w", file.Name, err)
		}
	}

	return nil
}

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
