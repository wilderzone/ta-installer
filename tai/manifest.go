package main

import (
	"io/fs"
	"os"
)

type Asset struct {
	// The unique name of the asset.
	Name        string `json:"name"`
	// The location of the asset's remote source.
	Source      string `json:"source"`
	// The location of the asset's local destination.
	Destination string `json:"destination"`
	// The files contained within the asset.
	Files       []File `json:"files"`
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
