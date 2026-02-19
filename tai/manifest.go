package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
)

type AssetType string
const (
	CORE         AssetType = "core"
	DLC          AssetType = "dlc"
	PREREQUISITE AssetType = "prerequisite"
)

type Asset struct {
	// The unique name of the asset.
	Name        string    `json:"name"`
	// The asset type.
	Type        AssetType `json:"type"`
	// The location of the asset's remote source.
	Source      string    `json:"source"`
	// The location of the asset's local destination.
	Destination string    `json:"destination"`
	// The files contained within the asset.
	Files       []File    `json:"files"`
}

type File struct {
	// The file's checksum.
	Checksum string `json:"checksum"`
	// The file name (including relative path & extension).
	File     string `json:"file"`
}

type Manifest struct {
	Assets  []Asset `json:"assets"`
	Version string  `json:"version"`
}

// Create a manifest asset from a given directory.
func CreateAsset(path string, name string) (Asset, error) {
	asset := Asset{}
	asset.Name = name

	f := os.DirFS(path)
	err := fs.WalkDir(f, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.IsDir() {
			data, err := fs.ReadFile(f, path)
			if err != nil {
				return err
			}
			checksum := Checksum(data)
			file := File{
				Checksum: checksum,
				File: path,
			}
			asset.Files = append(asset.Files, file)
		}

		return nil
	})

	return asset, err
}

// Create a manifest file from a given collection of asset files.
func CreateManifest(path string) (Manifest, error) {
	manifest := Manifest{}
	manifest.Version = "0.0.0"

	f := os.DirFS(path)
	err := fs.WalkDir(f, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.IsDir() {
			file, err := f.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			asset := Asset{}
			json.NewDecoder(file).Decode(&asset)
			manifest.Assets = append(manifest.Assets, asset)
		}

		return nil
	})

	return manifest, err
}

// Retrieve a manifest file from the given repo URL.
func GetManifest(repo string) (Manifest, error) {
	manifest := Manifest{}
	path, err := url.JoinPath(repo, "manifest.json")
	if err != nil {
		return manifest, err
	}

	fmt.Println("Retrieving manifest from:", path)
	response, err := http.Get(path)
	if err != nil {
		return manifest, err
	}
	defer response.Body.Close()

	fmt.Println("Status:", response.Status)
	err = json.NewDecoder(response.Body).Decode(&manifest)
	if err != nil {
		return manifest, err
	}

	return manifest, nil
}
