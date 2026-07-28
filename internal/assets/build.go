package assets

import (
	"encoding/json"
	"io/fs"
	"os"
)

// CreateAsset builds a manifest asset from a given directory, checksumming every file.
func CreateAsset(path string, name string) (Asset, error) {
	asset := Asset{}
	asset.Name = name
	asset.Dependencies = []string{}

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
			checksum, err := checksumReader(source)
			if err != nil {
				return err
			}
			file := File{
				Checksum: checksum,
				File:     path,
			}
			asset.Files = append(asset.Files, file)
		}

		return nil
	})

	return asset, err
}

// CreateManifest builds a manifest from a given collection of asset files.
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
