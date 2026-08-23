package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"wilderzone.org/tai/internal/assets"
)

// Handle the "uninstall" command.
func Uninstall(ctx context.Context, logger *Logger, arguments arguments) error {
	filterProvided := len(arguments.remove) != 0 || len(arguments.keep) != 0

	installed, err := readManifest(logger, arguments.path)
	if errors.Is(err, fs.ErrNotExist) {
		// No manifest means no installation. This is an expected outcome.
		logger.log(INSTALL_NOT_FOUND)
		return nil
	}
	if err != nil {
		// The manifest is present but unreadable.
		if filterProvided {
			// A surgical uninstall needs the manifest to know which files to remove.
			return userErrorf(err, "Can't remove a subset of the installed assets: the installation's manifest at %s is unreadable.", arguments.path)
		}
		// A full uninstall can still proceed, but confirm first.
		logger.log(MANIFEST_UNREADABLE, fmt.Sprintf("The installation's manifest at %s couldn't be read.", arguments.path))
		proceed := arguments.yes || logger.prompt(PROMPT_UNINSTALL_ANYWAY)
		if !proceed {
			return nil
		}
		return removeInstallation(ctx, logger, arguments.path)
	}

	// Determine which assets to remove. `--remove` maps to include and `--keep` to
	// exclude, so assets.FilterAssets yields the delete-set for every combination (and no
	// filter selects every asset).
	// Clone the result: with no filter, assets.FilterAssets returns its input, and the appends
	// below would otherwise write through to the manifest's own assets.
	toRemove := slices.Clone(assets.FilterAssets(installed.Assets, arguments.remove, arguments.keep))
	if filterProvided && len(toRemove) == 0 {
		return userErrorf(nil, "No installed assets matched the provided filters. Nothing will be uninstalled.")
	}

	// An asset left without something it depends on wouldn't work, so anything that
	// requires a removed asset has to be removed as well.
	dependents := assets.DependentsOf(installed.Assets, toRemove)
	toRemove = append(toRemove, dependents...)

	// Assets that were only installed to satisfy a dependency are pruned once nothing
	// needs them any more.
	survivors, orphans := assets.PruneOrphans(assets.DifferenceByName(installed.Assets, toRemove))
	toRemove = append(toRemove, orphans...)

	if len(dependents) != 0 {
		logger.log(UNINSTALL_DEPENDENTS, assets.AssetNames(dependents))
	}
	if len(orphans) != 0 {
		logger.log(UNINSTALL_ORPHANS, assets.AssetNames(orphans))
	}

	// If nothing would remain installed, remove the entire installation.
	if len(survivors) == 0 {
		if filterProvided {
			logger.log(UNINSTALL_MATCHED_ALL)
		}
		logger.log(UNINSTALL_EVERYTHING)
		proceed := arguments.yes || logger.prompt(PROMPT_PROCEED)
		if !proceed {
			return userErrorf(nil, "Uninstall cancelled.")
		}
		return removeInstallation(ctx, logger, arguments.path)
	}

	// Confirm before removing assets the filters didn't name. Orphaned dependencies
	// aren't confirmed - they were TAI's own doing, not the user's.
	if len(dependents) != 0 {
		proceed := arguments.yes || logger.prompt(PROMPT_PROCEED)
		if !proceed {
			return userErrorf(nil, "Uninstall cancelled.")
		}
	}

	// Surgically remove only the selected assets' files.
	logger.log(UNINSTALLING)
	survivorFiles := assetFilePaths(arguments.path, survivors)
	counter := logger.count(len(toRemove), "Removing")
	defer counter.done()
	failed := 0
	prune := map[string]bool{}
	for _, asset := range toRemove {
		// Stop before rewriting the manifest: leaving it listing the assets that are
		// only half-removed is what lets another `uninstall` finish the job.
		if err := ctx.Err(); err != nil {
			return err
		}
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
	_, err = storeManifest(Manifest{Assets: survivors, Version: installed.Version}, arguments.path)
	if err != nil {
		return userErrorf(err, "Removed the assets, but couldn't update the manifest at %s.", arguments.path)
	}

	if failed == len(toRemove) {
		logger.log(UNINSTALL_FAILURE)
	} else if failed != 0 {
		err := removeUninstallerFromRegistry()
		logger.detail("remove uninstaller from registry", err)
		logger.log(UNINSTALL_PARTIAL, arguments.path)
	} else {
		err := removeUninstallerFromRegistry()
		logger.detail("remove uninstaller from registry", err)
		logger.log(UNINSTALL_SUCCESS)
	}
	return nil
}

// removeInstallation clears the entire install directory and the external
// shortcuts. Used when no subset is requested, or when nothing would remain
// installed after a filtered uninstall.
func removeInstallation(ctx context.Context, logger *Logger, root string) error {
	logger.log(UNINSTALLING)

	// Remove all entries in the specified directory.
	entries, err := os.ReadDir(root)
	if err != nil {
		return userErrorf(err, "Couldn't read the installation directory %s.", root)
	}

	// When the running CLI lives inside the installation's own utilities directory - as it
	// does when launched by the bundled uninstaller - Windows won't let it delete that
	// directory, because the CLI's own image is locked. Skip it and let the external
	// uninstaller remove it once this process exits. Run from anywhere else (such as the
	// GUI's temp copy), the directory is deleted along with everything else.
	skipUtilities := runningInsideUtilities(root)
	targets := make([]os.DirEntry, 0, len(entries))
	for _, entry := range entries {
		if skipUtilities && entry.Name() == INSTALLATION_UTILITIES_DIRECTORY {
			continue
		}
		targets = append(targets, entry)
	}

	counter := logger.count(len(targets), "Removing")
	defer counter.done()
	failed := 0
	for _, entry := range targets {
		// Whatever is left keeps the shortcuts and the Registry entry pointing at a
		// real installation, so those stay until the removal actually completes.
		if err := ctx.Err(); err != nil {
			return err
		}
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
		logger.detail("remove shortcuts", err)
	}

	if failed == 0 {
		err := removeUninstallerFromRegistry()
		logger.detail("remove uninstaller from registry", err)
		logger.log(UNINSTALL_SUCCESS)
	} else if failed == len(targets) {
		logger.log(UNINSTALL_FAILURE)
	} else {
		err := removeUninstallerFromRegistry()
		logger.detail("remove uninstaller from registry", err)
		logger.log(UNINSTALL_PARTIAL, root)
	}
	return nil
}

// runningInsideUtilities reports whether the currently running executable lives directly
// inside the installation's utilities directory. When it does, the CLI can't delete that
// directory, so removeInstallation leaves it for the external uninstaller to clean up.
func runningInsideUtilities(root string) bool {
	self, err := os.Executable()
	if err != nil {
		return false
	}
	if resolved, err := filepath.EvalSymlinks(self); err == nil {
		self = resolved
	}
	utilities := filepath.Join(root, INSTALLATION_UTILITIES_DIRECTORY)
	if resolved, err := filepath.EvalSymlinks(utilities); err == nil {
		utilities = resolved
	}
	return samePath(filepath.Dir(self), utilities)
}

// assetFilePaths returns the set of on-disk paths for every file in the given assets.
func assetFilePaths(root string, group []Asset) map[string]bool {
	paths := map[string]bool{}
	for _, asset := range group {
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
