package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
	Version uint32  `json:"version"`
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
	manifest.Version = 0

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

// Read the content of the local manifest file.
func ReadManifest(root string) (Manifest, error) {
	manifest := Manifest{}
	source := filepath.Join(root, "manifest.json")
	fmt.Println("Reading local manifest from:", source)
	file, err := os.OpenFile(source, os.O_RDONLY, 0777)
	if err != nil {
		return manifest, err
	}

	err = json.NewDecoder(file).Decode(&manifest)
	if err != nil {
		return manifest, err
	}

	return manifest, nil
}

// Retrieve a manifest file from the given repo URL.
func GetManifest(repo string) (Manifest, error) {
	manifest := Manifest{}
	source, err := url.JoinPath(repo, "manifest.json")
	if err != nil {
		return manifest, err
	}

	fmt.Println("Retrieving manifest from:", source)
	response, err := http.Get(source)
	if err != nil {
		return manifest, err
	}
	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(&manifest)
	if err != nil {
		return manifest, err
	}

	return manifest, nil
}

// Download a manifest file from the given repo URL.
func DownloadManifest(repo string, root string) (Manifest, error) {
	manifest := Manifest{}
	destination := filepath.Join(root, "manifest.json")
	file, err := os.OpenFile(destination, os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return manifest, err
	}
	source, err := url.JoinPath(repo, "manifest.json")
	if err != nil {
		return manifest, err
	}

	fmt.Println("Downloading manifest from:", source)
	response, err := http.Get(source)
	if err != nil {
		return manifest, err
	}
	defer response.Body.Close()

	_, err = io.Copy(file, response.Body)
	file.Close()

	err = json.NewDecoder(response.Body).Decode(&manifest)
	if err != nil {
		return manifest, err
	}

	return manifest, nil
}

// Save a manifest file to disk.
func StoreManifest(manifest Manifest, root string) (Manifest, error) {
	destination := filepath.Join(root, "manifest.json")
	file, err := os.OpenFile(destination, os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return manifest, err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")
	err = encoder.Encode(manifest)
	if err != nil {
		return manifest, err
	}

	return manifest, nil
}
