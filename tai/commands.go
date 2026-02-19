package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Handle the "asset" command.
func GenerateAssetFile(arguments arguments) (string, error) {
	asset, err := CreateAsset(arguments.path, arguments.name)
	if err != nil {
		return "", errors.New("Failed to create asset file: " + err.Error())
	}
	path := filepath.Join(arguments.path, arguments.name+".json")
	file, err := os.OpenFile(path, os.O_CREATE, 0777)
	if err != nil {
		return "", errors.New("Failed to create asset file: " + err.Error())
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")
	err = encoder.Encode(asset)
	if err != nil {
		return "", errors.New("Failed to encode asset file: " + err.Error())
	}
	plural := "file:"
	if len(asset.Files) != 1 {
		plural = "files:"
	}
	message := fmt.Sprint("Created asset containing", len(asset.Files), plural, "\n", path)
	return message, nil
}

// Handle the "manifest" command.
func GenerateManifestFile(arguments arguments) (string, error) {
	manifest, err := CreateManifest(arguments.path)
	if err != nil {
		return "", errors.New("Failed to create manifest file: " + err.Error())
	}
	path := filepath.Join(arguments.path, "manifest.json")
	file, err := os.OpenFile(path, os.O_CREATE, 0777)
	if err != nil {
		return "", errors.New("Failed to create manifest file: " + err.Error())
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")
	err = encoder.Encode(manifest)
	if err != nil {
		return "", errors.New("Failed to encode manifest file: " + err.Error())
	}
	plural := "asset:"
	if len(manifest.Assets) != 1 {
		plural = "assets:"
	}
	message := fmt.Sprint("Created manifest containing", len(manifest.Assets), plural, "\n", path)
	return message, nil
}

// Handle the "install" command.
func Install(arguments arguments) (string, error) {
	manifest, err := GetManifest(arguments.repo)
	if err != nil {
		return "", errors.New("Failed to retrieve manifest: " + err.Error())
	}

	// Download and unpack assets.
	failed := []Asset{}
	prerequisitesFailed := false
	for index, asset := range manifest.Assets {
		message := fmt.Sprintf("\x1b[2m[%d/%d]\x1b[0m ", index + 1, len(manifest.Assets))
		message += asset.Name
		err = Download(arguments.repo, arguments.path, asset)
		if err != nil {
			message += ": \x1b[31mFailed\x1b[0m: " + err.Error()
			failed = append(failed, asset)
			if asset.Type == PREREQUISITE {
				prerequisitesFailed = true
			}
		} else {
			message += ": \x1b[32mDownloaded\x1b[0m"
		}
		fmt.Println(message)
	}

	// Install prerequisites.
	if !prerequisitesFailed {
		fmt.Println("Installing prerequisites...")
		err = Prerequisites(arguments.path)
		if err != nil {
			return "", errors.New("Failed to install prerequisites: " + err.Error())
		}
	}

	message := "Installed."
	if len(failed) == len(manifest.Assets) {
		message = "Installation failed."
	} else if len(failed) != 0 {
		message = "Partially installed. You may need to run the \"verify\" command."
	}
	return message, nil
}
