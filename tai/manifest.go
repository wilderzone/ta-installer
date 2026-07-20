package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
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

	logger.log(DOWNLOADING_MANIFEST, source)
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

// filterAssets returns the assets that satisfy the include/exclude terms. A term
// matches an asset by its type (e.g. "dlc") or its exact name (e.g. "Foo Bar").
// Include is applied first (an empty include list means "include all"), then
// exclude. Matching is case-sensitive.
func filterAssets(assets []Asset, include []string, exclude []string) []Asset {
	if len(include) == 0 && len(exclude) == 0 {
		return assets
	}
	matches := func(asset Asset, terms []string) bool {
		return slices.Contains(terms, string(asset.Type)) || slices.Contains(terms, asset.Name)
	}
	filtered := []Asset{}
	for _, asset := range assets {
		included := len(include) == 0 || matches(asset, include)
		if included && !matches(asset, exclude) {
			filtered = append(filtered, asset)
		}
	}
	return filtered
}

// intersectByName returns the assets from `available` whose name matches one of
// the `installed` assets, preserving the order of `available`.
func intersectByName(available []Asset, installed []Asset) []Asset {
	names := make(map[string]bool, len(installed))
	for _, asset := range installed {
		names[asset.Name] = true
	}
	result := []Asset{}
	for _, asset := range available {
		if names[asset.Name] {
			result = append(result, asset)
		}
	}
	return result
}

// mergeAssets returns the union of the already-installed `existing` assets and the
// newly-installed `added` assets, keyed by name, so that repeated installs are
// additive. For a name present in both, the `added` definition wins (it was just
// downloaded). Existing assets keep their original order; genuinely new assets are
// appended.
func mergeAssets(existing []Asset, added []Asset) []Asset {
	addedByName := make(map[string]Asset, len(added))
	for _, asset := range added {
		addedByName[asset.Name] = asset
	}
	merged := []Asset{}
	seen := make(map[string]bool, len(existing)+len(added))
	for _, asset := range existing {
		if replacement, ok := addedByName[asset.Name]; ok {
			merged = append(merged, replacement)
		} else {
			merged = append(merged, asset)
		}
		seen[asset.Name] = true
	}
	for _, asset := range added {
		if !seen[asset.Name] {
			merged = append(merged, asset)
			seen[asset.Name] = true
		}
	}
	return merged
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
