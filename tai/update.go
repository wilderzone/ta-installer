package main

import (
	"errors"
	"fmt"
	"io/fs"
)

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
	// install is not re-expanded to the full catalogue. Resolving from there picks up
	// any dependency the new version has added, which would otherwise leave the
	// installation broken. The flags have to be re-applied because the remote manifest
	// doesn't carry them.
	installedAssets := intersectByName(remoteManifest.Assets, localManifest.Assets)
	closure, err := resolveDependencies(remoteManifest.Assets, installedAssets)
	if err != nil {
		return userErrorf(err, "The repository at %s lists assets that can't be installed together. Please report this to the repository's maintainer.", arguments.repo)
	}
	remoteManifest.Assets = applyAutomatic(closure, explicitAssets(localManifest.Assets), localManifest.Assets)

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
		// Update the uninstaller's Registry entry.
		err := addUninstallerToRegistry(arguments.path)
		logger.detail("add uninstaller to registry", err)
		logger.log(UPDATE_SUCCESS)
	}
	return nil
}
