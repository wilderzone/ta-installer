package main

import (
	"archive/zip"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/uuid"
)

func Download(repo string, root string, asset Asset) error {
	// Download the asset into the temp directory.
	os.MkdirAll(filepath.Join(root, "temp"), 0777)
	id := uuid.New().String()
	if len(id) == 0 {
		return errors.New("Failed to generate UUID.")
	}
	tempPath := filepath.Join(root, "temp", id)
	file, err := os.OpenFile(tempPath, os.O_CREATE|os.O_TRUNC, 0777)
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
	defer response.Body.Close()
	_, err = io.Copy(file, response.Body)
	file.Close()

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
	os.MkdirAll(filepath.Join(root, asset.Destination), 0777)
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		path := filepath.Join(root, asset.Destination, file.Name)

		// Handle directories.
		if file.FileInfo().IsDir() {
			err = os.MkdirAll(path, 0777)
			if err != nil {
				return err
			}
			continue
		}

		// Open the destination file.
		output, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC, 0777)
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
