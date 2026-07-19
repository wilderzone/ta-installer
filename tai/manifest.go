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
			source, err := f.Open(path)
			if err != nil {
				return err
			}
			defer source.Close()
			checksum, err := ChecksumReader(source)
			if err != nil {
				return err
			}
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
func ReadManifest(logger *Logger, root string) (Manifest, error) {
	manifest := Manifest{}
	source := filepath.Join(root, MANIFEST_NAME)
	logger.log(READING_LOCAL_MANIFEST, source)
	file, err := os.OpenFile(source, os.O_RDONLY, os.FileMode(0777))
	if err != nil {
		// %w keeps fs.ErrNotExist detectable by callers via errors.Is.
		return manifest, fmt.Errorf("open manifest %s: %w", source, err)
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&manifest)
	if err != nil {
		return manifest, fmt.Errorf("decode manifest %s: %w", source, err)
	}

	return manifest, nil
}

// Retrieve a manifest file from the given repo URL.
func GetManifest(logger *Logger, repo string) (Manifest, error) {
	manifest := Manifest{}
	source, err := url.JoinPath(repo, MANIFEST_NAME)
	if err != nil {
		return manifest, fmt.Errorf("build manifest URL: %w", err)
	}

	logger.log(RETRIEVING_MANIFEST, source)
	response, err := http.Get(source)
	if err != nil {
		return manifest, fmt.Errorf("request %s: %w", source, err)
	}
	if response.StatusCode != http.StatusOK {
		return manifest, fmt.Errorf("request %s: unexpected status %s", source, response.Status)
	}
	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(&manifest)
	if err != nil {
		return manifest, fmt.Errorf("decode manifest from %s: %w", source, err)
	}

	return manifest, nil
}

// Download a manifest file from the given repo URL.
func DownloadManifest(logger *Logger, repo string, root string) (Manifest, error) {
	manifest := Manifest{}
	destination := filepath.Join(root, MANIFEST_NAME)
	source, err := url.JoinPath(repo, MANIFEST_NAME)
	if err != nil {
		return manifest, fmt.Errorf("build manifest URL: %w", err)
	}

	logger.log(DOWNLOADING_MANIFEST, source)
	response, err := http.Get(source)
	if err != nil {
		return manifest, fmt.Errorf("request %s: %w", source, err)
	}
	if response.StatusCode != http.StatusOK {
		return manifest, fmt.Errorf("request %s: unexpected status %s", source, response.Status)
	}
	defer response.Body.Close()

	// Read the response.
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return manifest, fmt.Errorf("read manifest from %s: %w", source, err)
	}

	// Unmarshal the JSON payload into the manifest struct.
	err = json.Unmarshal(body, &manifest)
	if err != nil {
		return manifest, fmt.Errorf("decode manifest from %s: %w", source, err)
	}

	// Create the manifest file.
	err = os.WriteFile(destination, body, os.FileMode(0644))
	if err != nil {
		return manifest, fmt.Errorf("write manifest to %s: %w", destination, err)
	}

	return manifest, nil
}

// Save a manifest file to disk.
func StoreManifest(manifest Manifest, root string) (Manifest, error) {
	destination := filepath.Join(root, MANIFEST_NAME)
	file, err := os.Create(destination)
	if err != nil {
		return manifest, fmt.Errorf("create manifest %s: %w", destination, err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")
	err = encoder.Encode(manifest)
	if err != nil {
		return manifest, fmt.Errorf("encode manifest %s: %w", destination, err)
	}

	return manifest, nil
}
