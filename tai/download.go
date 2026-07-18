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

	"github.com/google/uuid"
)

func Download(repo string, root string, asset Asset) error {
	// Download the asset into the temp directory.
	err := os.MkdirAll(filepath.Join(root, "temp"), os.FileMode(0777))
	if err != nil {
		return err
	}
	id := uuid.New().String()
	if len(id) == 0 {
		return errors.New("Failed to generate UUID.")
	}
	tempPath := filepath.Join(root, "temp", id)
	file, err := os.Create(tempPath)
	if err != nil {
		return err
	}
	source, err := url.JoinPath(repo, asset.Source)
	if err != nil {
		return err
	}
	response, err := http.Get(source)
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status %s for request to \"%s\".", response.Status, source)
	}
	defer response.Body.Close()
	_, err = io.Copy(file, response.Body)
	file.Close()
	if err != nil {
		return err
	}

	// Unpack the downloaded file into the destination directory.
	err = Unpack(root, tempPath, asset)
	if err != nil {
		return err
	}
	err = os.Remove(tempPath)
	if err != nil {
		return err
	}

	return nil
}

func Unpack(root string, archivePath string, asset Asset) error {
	destination := filepath.Join(root, asset.Destination)
	os.MkdirAll(destination, os.FileMode(0777))
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
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
				return err
			}
			continue
		}

		// Open the destination file.
		output, err := os.Create(path)
		if err != nil {
			return err
		}

		// Open the archived file.
		archived, err := file.Open()
		if err != nil {
			output.Close()
			return err
		}

		// Copy the file content.
		_, err = io.Copy(output, archived)
		output.Close()
		archived.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

func Prerequisites(root string) error {
	prerequisite1 := filepath.Join(root, "Binaries", "Redist", "directx_Jun2010_redist.exe")
	command1 := exec.Command(prerequisite1, "/T:" + filepath.Join(root, "temp", "dx"), "/Q")
	err := command1.Run()
	if err != nil {
		return err
	}

	prerequisite2 := filepath.Join(root, "temp", "dx", "DXSETUP.exe")
	command2 := exec.Command(prerequisite2, "/silent")
	err = command2.Run()
	if err != nil {
		return err
	}

	return nil
}
