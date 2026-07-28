package main

import (
	"errors"
	"fmt"
	"io/fs"
	"sync"

	"wilderzone.org/tai/internal/assets"
)

// Handle the "verify" command.
func Verify(logger *Logger, arguments arguments) error {
	manifest, err := readManifest(logger, arguments.path)
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
			} else {
				// Update the uninstaller's Registry entry.
				err := addUninstallerToRegistry(arguments.path)
				logger.detail("add uninstaller to registry", err)
			}
		}
	}

	// Also verify the installation utilities (tai-cli.exe and friends), whose checksums are
	// fetched from the CDN, separately from the game manifest.
	verifyUtilities(logger, arguments)

	logger.log(INSTALL_VERIFIED)
	return nil
}

// checkAssets verifies each asset in the manifest against what is on disk,
// logging per-asset progress, and returns the assets that need to be
// re-downloaded. `damagedLabel` is the word shown for a damaged asset
// (e.g. "Damaged" for verify, "Needs patching" for update).
func checkAssets(logger *Logger, root string, manifest Manifest, damagedLabel string) []Asset {
	counter := logger.count(len(manifest.Assets), "Verifying")
	defer counter.done()
	var mu sync.Mutex
	damaged := []Asset{}

	parallel(manifest.Assets, VERIFY_CONCURRENCY, func(asset Asset) {
		results := assets.VerifyAsset(root, asset)
		message := asset.Name
		if len(results) == 0 {
			message += ": " + logger.colour("Verified", GREEN)
		} else {
			mu.Lock()
			damaged = append(damaged, asset)
			mu.Unlock()
			message += ": " + logger.colour(damagedLabel, RED) + " " + logger.colour(assets.Describe(results), GREY)
		}
		counter.next(message)
	})

	return damaged
}
