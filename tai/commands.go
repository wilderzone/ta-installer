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
	manifest, err := GetManifest(logger, arguments.repo)
	if err != nil {
		return userErrorf(err, "Couldn't download the game files from %s. Check your connection and that the repository URL is correct.", arguments.repo)
	}

	// Determine the subset of assets to install.
	subset := filterAssets(manifest.Assets, arguments.include, arguments.exclude)
	if len(subset) == 0 {
		return userErrorf(nil, "No assets matched the given filter.")
	}

	// Merge with any existing installation so repeated installs are additive.
	// Keep the existing (older) version so a later `update` still reconciles every
	// asset; a fresh install adopts the remote version.
	existing, existErr := ReadManifest(logger, arguments.path)
	if existErr == nil {
		manifest.Assets = mergeAssets(existing.Assets, subset)
		manifest.Version = min(existing.Version, manifest.Version)
	} else {
		manifest.Assets = subset
	}

	// Persist the manifest before downloading, so an interrupted install is still recognised.
	_, err = StoreManifest(manifest, arguments.path)
	if err != nil {
		return userErrorf(err, "Couldn't save the manifest to %s.", arguments.path)
	}

	// Download and unpack the newly-selected subset.
	failed := downloadAssets(logger, arguments, subset, "Downloaded")

	// Create the logs directory.
	err = os.MkdirAll(filepath.Join(arguments.path, LOGS_DIRECTORY), os.FileMode(0755))
	if err != nil {
		return fmt.Errorf("create logs directory: %w", err)
	}

	// Create shortcuts.
	err = createShortcuts(logger, arguments.path, arguments.shortcuts)
	if err != nil && arguments.shortcuts {
		logger.log(SHORTCUTS_FAILURE)
	}

	// Install prerequisites.
	err = Prerequisites(logger, arguments)
	if err != nil {
		return userErrorOr(err, "The game files were installed, but DirectX 9 couldn't be installed. The game may not start until DirectX 9 is present.")
	}

	if len(failed) == len(subset) {
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
	manifest.Assets = filterAssets(manifest.Assets, arguments.include, arguments.exclude)

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
	filterProvided := len(arguments.remove) != 0 || len(arguments.keep) != 0

	installed, err := ReadManifest(logger, arguments.path)
	if errors.Is(err, fs.ErrNotExist) {
		// No manifest means no installation. This is an expected outcome.
		logger.log(INSTALL_NOT_FOUND)
		return nil
	}
	if err != nil {
		// The manifest is present but unreadable.
		if filterProvided {
			// A surgical uninstall needs the manifest to know which files to remove.
			return userErrorf(err, "Can't remove a subset: the installation's manifest at %s is unreadable.", arguments.path)
		}
		// A full uninstall can still proceed, but confirm first.
		logger.log(MANIFEST_UNREADABLE, fmt.Sprintf("The installation's manifest at %s couldn't be read.", arguments.path))
		proceed := arguments.yes || logger.prompt(PROMPT_UNINSTALL_ANYWAY)
		if !proceed {
			return nil
		}
		return removeInstallation(logger, arguments.path)
	}

	// Determine which assets to remove. `--remove` maps to include and `--keep` to
	// exclude, so filterAssets yields the delete-set for every combination (and no
	// filter selects every asset).
	toRemove := filterAssets(installed.Assets, arguments.remove, arguments.keep)
	if filterProvided && len(toRemove) == 0 {
		return userErrorf(nil, "No installed assets matched — nothing to uninstall.")
	}

	// If nothing would remain installed, remove the entire installation.
	survivors := differenceByName(installed.Assets, toRemove)
	if len(survivors) == 0 {
		return removeInstallation(logger, arguments.path)
	}

	// Surgically remove only the selected assets' files.
	logger.log(UNINSTALLING)
	survivorFiles := assetFilePaths(arguments.path, survivors)
	counter := logger.count(len(toRemove))
	failed := 0
	prune := map[string]bool{}
	for _, asset := range toRemove {
		message := asset.Name
		err := removeAssetFiles(arguments.path, asset, survivorFiles, prune)
		if err != nil {
			message += ": " + logger.colour("Failed", RED) + ": " + err.Error()
			failed += 1
		} else {
			message += ": " + logger.colour("Removed", GREEN)
		}
		counter.next(message)
	}
	pruneEmptyDirectories(arguments.path, prune)

	// Rewrite the manifest so it reflects the assets that remain installed. Keep the
	// existing version - the surviving assets are unchanged.
	_, err = StoreManifest(Manifest{Assets: survivors, Version: installed.Version}, arguments.path)
	if err != nil {
		return userErrorf(err, "Removed the assets, but couldn't update the manifest at %s.", arguments.path)
	}

	if failed == len(toRemove) {
		logger.log(UNINSTALL_FAILURE)
	} else if failed != 0 {
		logger.log(UNINSTALL_PARTIAL)
	} else {
		logger.log(UNINSTALL_SUCCESS)
	}
	return nil
}

// removeInstallation clears the entire install directory and the external
// shortcuts. Used when no subset is requested, or when nothing would remain
// installed after a filtered uninstall.
func removeInstallation(logger *Logger, root string) error {
	logger.log(UNINSTALLING)

	// Remove all entries in the specified directory.
	entries, err := os.ReadDir(root)
	if err != nil {
		return userErrorf(err, "Couldn't read the installation directory %s.", root)
	}
	counter := logger.count(len(entries))
	failed := 0
	for _, entry := range entries {
		message := entry.Name()
		path := filepath.Join(root, entry.Name())
		err = os.RemoveAll(path)
		if err != nil {
			message += ": " + logger.colour("Failed", RED) + ": " + err.Error()
			failed += 1
		} else {
			message += ": " + logger.colour("Removed", GREEN)
		}
		counter.next(message)
	}

	// Remove the external Start Menu and Desktop shortcuts (best-effort).
	if err := removeShortcuts(root); err != nil {
		logger.log(SHORTCUTS_REMOVE_FAILURE)
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

// assetFilePaths returns the set of on-disk paths for every file in the given assets.
func assetFilePaths(root string, assets []Asset) map[string]bool {
	paths := map[string]bool{}
	for _, asset := range assets {
		for _, file := range asset.Files {
			paths[filepath.Join(root, asset.Destination, file.File)] = true
		}
	}
	return paths
}

// removeAssetFiles deletes an asset's files from disk, skipping any path still
// owned by a surviving asset. Parent directories of removed files are recorded in
// `prune` for later empty-directory cleanup. Returns the first error encountered.
func removeAssetFiles(root string, asset Asset, keep map[string]bool, prune map[string]bool) error {
	var firstErr error
	for _, file := range asset.Files {
		path := filepath.Join(root, asset.Destination, file.File)
		if keep[path] {
			continue // Shared with a surviving asset.
		}
		err := os.Remove(path)
		if err != nil && !errors.Is(err, fs.ErrNotExist) && firstErr == nil {
			firstErr = err
		}
		prune[filepath.Dir(path)] = true
	}
	return firstErr
}

// pruneEmptyDirectories removes now-empty directories (deepest first), climbing up
// toward `root` but never removing `root` itself. Best-effort: a non-empty directory
// fails os.Remove harmlessly and stops the climb.
func pruneEmptyDirectories(root string, dirs map[string]bool) {
	root = filepath.Clean(root)
	ordered := make([]string, 0, len(dirs))
	for dir := range dirs {
		ordered = append(ordered, filepath.Clean(dir))
	}
	// Deepest paths first, so children are removed before their parents.
	slices.SortFunc(ordered, func(a string, b string) int {
		return len(b) - len(a)
	})
	for _, start := range ordered {
		for dir := start; dir != root; dir = filepath.Dir(dir) {
			if os.Remove(dir) != nil {
				break
			}
		}
	}
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

	// Restrict the update to the assets that are actually installed, so a subset
	// install is not re-expanded to the full catalogue.
	remoteManifest.Assets = intersectByName(remoteManifest.Assets, localManifest.Assets)

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
