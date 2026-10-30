package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	"wilderzone.org/tai/internal/assets"
)

// Handle the "update" command.
func Update(ctx context.Context, logger *Logger, arguments arguments) error {
	logger.log(CHECKING_FOR_UPDATES)

	localManifest, err := readManifest(logger, arguments.path)
	if errors.Is(err, fs.ErrNotExist) {
		return userErrorf(err, "No installation found at %s. Run `install` first.", arguments.path)
	}
	if err != nil {
		return userErrorf(err, "The installation's manifest at %s couldn't be read.", arguments.path)
	}
	remoteManifest, err := getManifest(ctx, logger, arguments.repo)
	if err != nil {
		return userErrorf(err, "Couldn't check for updates from %s. Check your connection and that the repository URL is correct.", arguments.repo)
	}

	// Patching stages its downloads through the temp directory, so clear it on the way out
	// regardless of the return condition - including when the run is interrupted mid-download.
	defer clearTempDirectory(arguments.path)

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
	installedAssets := assets.IntersectByName(remoteManifest.Assets, localManifest.Assets)
	closure, err := assets.ResolveDependencies(remoteManifest.Assets, installedAssets)
	if err != nil {
		return userErrorf(err, "The repository at %s lists assets that can't be installed together. Please report this to the repository's maintainer.", arguments.repo)
	}
	remoteManifest.Assets = assets.ApplyAutomatic(closure, assets.ExplicitAssets(localManifest.Assets), localManifest.Assets)

	// Find what needs patching before saving the new manifest, so an update that stops here
	// - interrupted, or turned down for lack of space - is still found by the next `update`.
	logger.log(CHECKING_ASSETS)
	patch := checkAssets(ctx, logger, arguments.path, remoteManifest, "Needs patching")
	// Unchecked assets may need patching as well, so the update isn't complete.
	if err := ctx.Err(); err != nil {
		return err
	}
	if checkFreeSpace(logger, arguments.path, patch, "the update") {
		if arguments.stopIfShort {
			return stoppedShort()
		}
		proceed := arguments.yes || logger.prompt(PROMPT_PROCEED)
		if !proceed {
			return userErrorf(nil, "Update cancelled.")
		}
	}

	// Save the new manifest, and then patch the local assets.
	_, err = storeManifest(remoteManifest, arguments.path)
	if err != nil {
		return userErrorf(err, "Couldn't save the updated manifest to %s.", arguments.path)
	}
	if len(patch) == 0 {
		logger.log(UPDATE_SUCCESS)
		return nil
	}

	logger.log(PATCHING_ASSETS)
	failed := downloadAssets(ctx, logger, arguments, patch, "Patched")
	if err := ctx.Err(); err != nil {
		return err
	}

	if len(failed) == len(patch) {
		logger.log(UPDATE_FAILURE)
	} else if len(failed) != 0 {
		logger.log(UPDATE_PARTIAL)
	} else {
		// Update the uninstaller's Registry entry.
		err := addUninstallerToRegistry(arguments.path, true)
		logger.detail("add uninstaller to registry", err)
		logger.log(UPDATE_SUCCESS)
	}
	return nil
}
