package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// Handle the "install" command.
func Install(logger *Logger, arguments arguments) error {
	manifest, err := getManifest(logger, arguments.repo)
	if err != nil {
		return userErrorf(err, "Couldn't download the game files from %s. Check your internet connection and that the repository URL is correct.", arguments.repo)
	}

	// Determine the filtered subset of assets.
	explicit := filterAssets(manifest.Assets, arguments.include, arguments.exclude)
	if len(explicit) == 0 {
		return userErrorf(nil, "No assets matched the given filter.")
	}

	// Pull in the dependencies of those assets.
	closure, err := resolveDependencies(manifest.Assets, explicit)
	if err != nil {
		return userErrorf(err, "The repository at %s lists assets that can't be installed together. Please report this to the repository's maintainer.", arguments.repo)
	}
	required := differenceByName(closure, explicit)

	// `--include` only under-specifies the selection, so missing dependencies are added
	// to it. `--exclude` is the user explicitly ruling an asset out, so silently
	// installing it anyway would be wrong.
	for _, asset := range required {
		if matchesTerms(asset, arguments.exclude) {
			return userErrorf(nil, "Can't install %q: it requires %q, which --exclude omits. Remove %q from --exclude to continue.", dependentOf(closure, asset.Name), asset.Name, asset.Name)
		}
	}
	if len(required) != 0 {
		logger.log(INSTALLING_DEPENDENCIES, assetNames(required))
	}

	// Create the installation directory, if it doesn't already exist.
	err = os.MkdirAll(arguments.path, os.FileMode(0755))
	if err != nil {
		return userErrorf(err, "Couldn't create the target directory %s.", arguments.path)
	}

	// Merge with any existing installation so repeated installs are additive.
	// Keep the existing (older) version so a later `update` still reconciles every
	// asset; a fresh install adopts the remote version.
	// Only trust the existing manifest if it read cleanly - a half-decoded one would
	// misreport both what is installed and how it got there.
	existing, existErr := readManifest(logger, arguments.path)
	installed := []Asset{}
	if existErr == nil {
		installed = existing.Assets
		manifest.Assets = mergeAssets(installed, closure)
		manifest.Version = min(existing.Version, manifest.Version)
	} else {
		manifest.Assets = closure
	}
	manifest.Assets = applyAutomatic(manifest.Assets, explicit, installed)

	// Persist the manifest before downloading, so an interrupted install is still recognised.
	_, err = storeManifest(manifest, arguments.path)
	if err != nil {
		return userErrorf(err, "Couldn't save the manifest to %s.", arguments.path)
	}

	// Download the installation utilities.
	utilitiesErr := downloadInstallationUtilities(logger, arguments)
	if utilitiesErr != nil {
		// Installation utilities are non-essential, so shouldn't throw an actual error.
		logger.log(UTILITIES_FAILURE)
		logger.detail("download installation utilities", utilitiesErr)
	} else {
		// Create a shortcut in the root directory for convenience.
		err := createUninstallShortcut(arguments.path)
		logger.detail("create uninstall shortcut: %w", err)
		// Add the uninstall utility to the Registry so the game can be uninstalled through Windows.
		err = addUninstallerToRegistry(arguments.path)
		logger.detail("add uninstaller to registry", err)
		if err != nil {
			logger.log(UTILITIES_FAILURE)
		}
	}

	logger.log(INSTALLING)

	// Download the selected assets, plus any dependency that isn't installed yet. An
	// already-installed dependency is left alone - reinstalling it could mean
	// re-downloading the entire game just to add a small asset, and repairing a damaged
	// asset is `verify`'s job.
	requested := make(map[string]bool, len(explicit))
	for _, asset := range explicit {
		requested[asset.Name] = true
	}
	present := make(map[string]bool, len(installed))
	for _, asset := range installed {
		present[asset.Name] = true
	}
	subset := []Asset{}
	for _, asset := range closure {
		if requested[asset.Name] || !present[asset.Name] {
			subset = append(subset, asset)
		}
	}
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
		logger.detail("create shortcuts", err)
	}

	// Install prerequisites.
	err = prerequisites(logger, arguments)
	if err != nil {
		return userErrorOr(err, "The game files were installed, but DirectX 9 couldn't be installed. The game may not start until DirectX 9 is present.")
	}

	// Recalculate the size of the installation.
	if utilitiesErr == nil {
		err := addUninstallerToRegistry(arguments.path)
		logger.detail("add uninstaller to registry", err)
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
