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
