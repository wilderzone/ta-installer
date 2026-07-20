package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"text/tabwriter"
)

// downloadAssets downloads and unpacks each asset concurrently, logging per-asset
// progress with the given success label, and returns the assets that failed.
func downloadAssets(logger *Logger, arguments arguments, assets []Asset, successLabel string) []Asset {
	counter := logger.count(len(assets))
	var mu sync.Mutex
	failed := []Asset{}
	parallel(assets, DOWNLOAD_CONCURRENCY, func(asset Asset) {
		message := asset.Name
		err := Download(arguments.repo, arguments.path, asset)
		if err != nil {
			message += ": " + logger.colour("Failed", RED) + ": " + err.Error()
			mu.Lock()
			failed = append(failed, asset)
			mu.Unlock()
		} else {
			message += ": " + logger.colour(successLabel, GREEN)
		}
		counter.next(message)
	})
	return failed
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
	logger.log(ASSET_CREATED, "Created asset containing", len(asset.Files), logger.plural(len(asset.Files), "file: ", "files: "), path)
	return nil
}

// Handle the "manifest" command.
func GenerateManifestFile(logger *Logger, arguments arguments) error {
	manifest, err := CreateManifest(arguments.path)
	if err != nil {
		return userErrorf(err, "Couldn't read the asset files in %s to build the manifest.", arguments.path)
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
	logger.log(MANIFEST_CREATED, "Created manifest containing", len(manifest.Assets), logger.plural(len(manifest.Assets), "asset: ", "assets: "), path)
	return nil
}

// Handle the "check" command.
func Check(logger *Logger, arguments arguments) error {
	_, err := ReadManifest(logger, arguments.path)
	if errors.Is(err, fs.ErrNotExist) {
		// No manifest means no installation. This is an expected outcome.
		logger.log(INSTALL_NOT_FOUND)
		return nil
	}
	if err != nil {
		// The manifest exists but couldn't be read - a genuine problem.
		return userErrorf(err, "Found an installation at %s, but its manifest couldn't be read.", arguments.path)
	}

	logger.log(INSTALL_FOUND)
	return nil
}

// Handle the "check-repo" command.
func CheckRepo(logger *Logger, arguments arguments) error {
	_, err := GetManifest(logger, arguments.repo)
	if err != nil {
		logger.log(REPO_NOT_FOUND)
		return nil
	}

	logger.log(REPO_FOUND)
	return nil
}

// Handle the "install" command.
func Install(logger *Logger, arguments arguments) error {
	manifest, err := DownloadManifest(logger, arguments.repo, arguments.path)
	if err != nil {
		return userErrorf(err, "Couldn't download the game files from %s. Check your connection and that the repository URL is correct.", arguments.repo)
	}

	// Download and unpack assets.
	failed := downloadAssets(logger, arguments, manifest.Assets, "Downloaded")

	// Create the logs directory.
	err = os.MkdirAll(filepath.Join(arguments.path, LOGS_DIRECTORY), os.FileMode(0755))
	if err != nil {
		return fmt.Errorf("create logs directory: %w", err)
	}

	// Create shortcuts.
	if arguments.shortcuts {
		err := createShortcuts(logger, arguments.path)
		if err != nil {
			return userErrorf(err, "Failed to create shortcuts.")
		}
	}

	// Install prerequisites.
	err = Prerequisites(logger, arguments)
	if err != nil {
		return userErrorOr(err, "The game files were installed, but DirectX 9 couldn't be installed. The game may not start until DirectX 9 is present.")
	}

	if len(failed) == len(manifest.Assets) {
		logger.log(INSTALL_FAILURE)
	} else if len(failed) != 0 {
		logger.log(INSTALL_PARTIAL)
	} else {
		logger.log(INSTALL_SUCCESS)
	}

	// Remove the temp directory. This is just courtesy cleanup - it's ok if it fails.
	defer os.RemoveAll(filepath.Join(arguments.path, TEMP_DIRECTORY))

	return nil
}

// Handle the "list" command.
func List(logger *Logger, arguments arguments) error {
	if logger.silent {
		return nil
	}

	var manifest Manifest
	var err error

	if len(arguments.path) != 0 {
		// List manifest assets in the specified directory.
		manifest, err = ReadManifest(logger, arguments.path)
		if errors.Is(err, fs.ErrNotExist) {
			logger.log(INSTALL_NOT_FOUND)
			return nil
		}
		if err != nil {
			return userErrorf(err, "The installation at %s has a manifest that couldn't be read.", arguments.path)
		}
	} else if len(arguments.repo) != 0 {
		// List manifest assets in the specified repo.
		manifest, err = GetManifest(logger, arguments.repo)
		if err != nil {
			logger.log(REPO_NOT_FOUND)
			return nil
		}
	}

	// Filter the manifest's assets based on the provided include/exclude arguments.
	// Include is applied first, then exclude. Empty include means include all.
	// Empty exclude means exclude nothing.
	if len(arguments.include) != 0 || len(arguments.exclude) != 0 {
		var included = []Asset{}
		for _, asset := range manifest.Assets {
			t := string(asset.Type)
			include := len(arguments.include) == 0 || slices.Contains(arguments.include, t)
			exclude := slices.Contains(arguments.exclude, t)
			if include && !exclude {
				included = append(included, asset)
			}
		}
		manifest.Assets = included
	}

	// Output the manifest's contents.
	if logger.machine {
		data, err := json.Marshal(manifest)
		if err != nil {
			logger.log(READING_MANIFEST_FAILED, err.Error())
		}
		logger.log(MANIFEST_CONTENTS, string(data))
	} else {
		fmt.Printf("Version: v%d\n", manifest.Version)
		w := tabwriter.NewWriter(os.Stdout, 1, 1, 3, ' ', 0)
		fmt.Fprint(w, "Name\tType\tFiles\t\n")
		fmt.Fprint(w, "----\t----\t-----\t\n")
		for _, asset := range manifest.Assets {
			fmt.Fprintf(w, "%s\t%s\t%d\t\n", asset.Name, asset.Type, len(asset.Files))
		}
		w.Flush()
	}

	return nil
}

// Handle the "uninstall" command.
func Uninstall(logger *Logger, arguments arguments) error {
	_, err := ReadManifest(logger, arguments.path)
	if errors.Is(err, fs.ErrNotExist) {
		// No manifest means no installation. This is an expected outcome.
		logger.log(INSTALL_NOT_FOUND)
		return nil
	}
	if err != nil {
		// The manifest is present but unreadable. We can still remove the
		// installation, but confirm first since we can't verify what we're deleting.
		logger.log(MANIFEST_UNREADABLE, fmt.Sprintf("The installation's manifest at %s couldn't be read.", arguments.path))
		proceed := arguments.yes || logger.prompt(PROMPT_UNINSTALL_ANYWAY)
		if !proceed {
			return nil
		}
	}

	logger.log(UNINSTALLING)

	// Remove all entries in the specified directory.
	entries, err := os.ReadDir(arguments.path)
	if err != nil {
		return userErrorf(err, "Couldn't read the installation directory %s.", arguments.path)
	}
	counter := logger.count(len(entries))
	failed := 0
	for _, entry := range entries {
		message := entry.Name()
		path := filepath.Join(arguments.path, entry.Name())
		err = os.RemoveAll(path)
		if err != nil {
			message += ": " + logger.colour("Failed", RED) + ": " + err.Error()
			failed += 1
		} else {
			message += ": " + logger.colour("Removed", GREEN)
		}
		counter.next(message)
	}

	if failed == len(entries) {
		logger.log(UNINSTALL_FAILURE)
	} else if failed != 0 {
		logger.log(UNINSTALL_PARTIAL)
	} else {
		logger.log(UNINSTALL_SUCCESS)
	}
	return nil
}

// Handle the "update" command.
func Update(logger *Logger, arguments arguments) error {
	logger.log(CHECKING_FOR_UPDATES)

	localManifest, err := ReadManifest(logger, arguments.path)
	if errors.Is(err, fs.ErrNotExist) {
		return userErrorf(err, "No installation found at %s. Run `install` first.", arguments.path)
	}
	if err != nil {
		return userErrorf(err, "The installation's manifest at %s couldn't be read.", arguments.path)
	}
	remoteManifest, err := GetManifest(logger, arguments.repo)
	if err != nil {
		return userErrorf(err, "Couldn't check for updates from %s. Check your connection and that the repository URL is correct.", arguments.repo)
	}

	// Compare the version fields of the local and remote manifests.
	if localManifest.Version >= remoteManifest.Version {
		version := logger.colour(fmt.Sprintf("v%d", localManifest.Version), GREEN)
		logger.log(INSTALL_UP_TO_DATE, "Installation is up to date: " + version)
		
		return nil
	}
	old := logger.colour(fmt.Sprintf("v%d", localManifest.Version), RED)
	new := logger.colour(fmt.Sprintf("v%d", remoteManifest.Version), GREEN)
	logger.log(UPDATE_FOUND, fmt.Sprintf("%s → %s\n", old, new))

	// If the remote version is greater, save the new manifest, and then patch the local assets.
	manifest, err := StoreManifest(remoteManifest, arguments.path)
	if err != nil {
		return userErrorf(err, "Couldn't save the updated manifest to %s.", arguments.path)
	}

	logger.log(CHECKING_ASSETS)
	patch := checkAssets(logger, arguments.path, manifest, "Needs patching")
	if len(patch) == 0 {
		logger.log(UPDATE_SUCCESS)
		return nil
	}

	logger.log(PATCHING_ASSETS)
	failed := downloadAssets(logger, arguments, patch, "Patched")

	if len(failed) == len(patch) {
		logger.log(UPDATE_FAILURE)
	} else if len(failed) != 0 {
		logger.log(UPDATE_PARTIAL)
	} else {
		logger.log(UPDATE_SUCCESS)
	}
	return nil
}

// Handle the "verify" command.
func Verify(logger *Logger, arguments arguments) error {
	manifest, err := ReadManifest(logger, arguments.path)
	if errors.Is(err, fs.ErrNotExist) {
		return userErrorf(err, "No installation found at %s to verify.", arguments.path)
	}
	if err != nil {
		return userErrorf(err, "The installation's manifest at %s couldn't be read.", arguments.path)
	}

	damaged := checkAssets(logger, arguments.path, manifest, "Damaged")

	if len(damaged) != 0 {
		message := logger.plural(len(damaged), "asset is damaged and needs", "assets are damaged and need") + " to be redownloaded. "
		logger.log(ASSETS_ARE_DAMAGED, fmt.Sprintf("%d %s", len(damaged), message))

		proceed := arguments.yes || logger.prompt(PROMPT_PROCEED)
		if proceed {
			logger.log(REDOWNLOADING)
			failed := downloadAssets(logger, arguments, damaged, "Downloaded")
			if len(failed) != 0 {
				return userErrorf(nil, "Some assets couldn't be re-downloaded. Run `verify` again to retry.")
			}
		}
	}

	logger.log(INSTALL_VERIFIED)
	return nil
}
