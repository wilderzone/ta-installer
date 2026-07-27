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
	Name         string    `json:"name"`
	// The asset type.
	Type         AssetType `json:"type"`
	// The location of the asset's remote source.
	Source       string    `json:"source"`
	// The location of the asset's local destination.
	Destination  string    `json:"destination"`
	// The names of any other assets that this asset depends on.
	Dependencies []string  `json:"dependencies"`
	// Whether this asset was installed to satisfy another asset's dependency, rather
	// than being asked for directly. Only ever set in a local manifest.
	Automatic    bool      `json:"automatic,omitempty"`
	// The files contained within the asset.
	Files        []File    `json:"files"`
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

// Handle the "asset" command.
func GenerateAssetFile(logger *Logger, arguments arguments) error {
	asset, err := CreateAsset(arguments.path, arguments.name)
	if err != nil {
		return userErrorf(err, "Couldn't read the files in %s to build the asset.", arguments.path)
	}
	path := filepath.Join(arguments.path, arguments.name + ".json")
	file, err := os.Create(path)
	if err != nil {
		return userErrorf(err, "Couldn't create the asset file at %s.", path)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")
	err = encoder.Encode(asset)
	if err != nil {
		return userErrorf(err, "Couldn't write the asset file at %s.", path)
	}
	message := fmt.Sprintf("Created asset containing %d %s: %s", len(asset.Files), logger.plural(len(asset.Files), "file", "files"), path)
	logger.log(ASSET_CREATED, message)
	return nil
}

// Handle the "manifest" command.
func GenerateManifestFile(logger *Logger, arguments arguments) error {
	manifest, err := CreateManifest(arguments.path)
	if err != nil {
		return userErrorf(err, "Couldn't read the asset files in %s to build the manifest.", arguments.path)
	}
	// Catch a broken dependency graph here, rather than mid-install.
	err = validateDependencies(manifest.Assets)
	if err != nil {
		return userErrorf(err, "The asset files in %s have a dependency problem: %s", arguments.path, err.Error())
	}
	path := filepath.Join(arguments.path, MANIFEST_NAME)
	file, err := os.Create(path)
	if err != nil {
		return userErrorf(err, "Couldn't create the manifest file at %s.", path)
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")
	err = encoder.Encode(manifest)
	if err != nil {
		return userErrorf(err, "Couldn't write the manifest file at %s.", path)
	}
	message := fmt.Sprintf("Created manifest containing %d %s: %s", len(manifest.Assets), logger.plural(len(manifest.Assets), "asset", "assets"), path)
	logger.log(MANIFEST_CREATED, message)
	return nil
}

// getUtilitiesAsset fetches the installation utilities' checksum file from the given repo
// and returns it as an Asset ready to verify against the installed files. The published
// file (generated by the `asset` command at release time) carries only names and checksums,
// so the Source and Destination - which the caller needs to locate and re-download the
// utilities - are filled in here from the known INSTALLATION_UTILITIES values.
func getUtilitiesAsset(repo string) (Asset, error) {
	asset := Asset{}
	source, err := url.JoinPath(repo, UTILITIES_CHECKSUMS_SOURCE)
	if err != nil {
		return asset, fmt.Errorf("build utilities checksums URL: %w", err)
	}

	response, err := http.Get(source)
	if err != nil {
		return asset, fmt.Errorf("request %s: %w", source, err)
	}
	if response.StatusCode != http.StatusOK {
		return asset, fmt.Errorf("request %s: unexpected status %s", source, response.Status)
	}
	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(&asset)
	if err != nil {
		return asset, fmt.Errorf("decode utilities checksums from %s: %w", source, err)
	}

	asset.Name = INSTALLATION_UTILITIES.Name
	asset.Source = INSTALLATION_UTILITIES.Source
	asset.Destination = INSTALLATION_UTILITIES.Destination
	return asset, nil
}

// matchesTerms reports whether any of the given filter terms matches the asset, by
// its type (e.g. "dlc") or its exact name (e.g. "Foo Bar"). Matching is case-sensitive.
func matchesTerms(asset Asset, terms []string) bool {
	return slices.Contains(terms, string(asset.Type)) || slices.Contains(terms, asset.Name)
}

// filterAssets returns the assets that satisfy the include/exclude terms. Include is
// applied first (an empty include list means "include all"), then exclude.
func filterAssets(assets []Asset, include []string, exclude []string) []Asset {
	if len(include) == 0 && len(exclude) == 0 {
		return assets
	}
	filtered := []Asset{}
	for _, asset := range assets {
		included := len(include) == 0 || matchesTerms(asset, include)
		if included && !matchesTerms(asset, exclude) {
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

// differenceByName returns the assets from `all` whose name is not among the
// `removed` assets, preserving the order of `all`.
func differenceByName(all []Asset, removed []Asset) []Asset {
	names := make(map[string]bool, len(removed))
	for _, asset := range removed {
		names[asset.Name] = true
	}
	result := []Asset{}
	for _, asset := range all {
		if !names[asset.Name] {
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

// applyAutomatic stamps the `Automatic` flag onto each asset, returning a new slice.
// An asset named in `explicit` was asked for directly, so it is never automatic (which
// also promotes an asset that was previously pulled in as a dependency); an asset
// already recorded in `previous` keeps the flag it had; anything else is new and only
// present to satisfy a dependency.
//
// Repository manifests never carry the flag, and mergeAssets lets the freshly
// downloaded definition win, so it has to be re-applied whenever the local manifest is
// rewritten.
func applyAutomatic(assets []Asset, explicit []Asset, previous []Asset) []Asset {
	requested := make(map[string]bool, len(explicit))
	for _, asset := range explicit {
		requested[asset.Name] = true
	}
	installed := make(map[string]bool, len(previous))
	for _, asset := range previous {
		installed[asset.Name] = asset.Automatic
	}

	stamped := make([]Asset, 0, len(assets))
	for _, asset := range assets {
		if requested[asset.Name] {
			asset.Automatic = false
		} else {
			automatic, known := installed[asset.Name]
			asset.Automatic = !known || automatic
		}
		stamped = append(stamped, asset)
	}
	return stamped
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
