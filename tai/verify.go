package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"sync"

	"wilderzone.org/tai/internal/assets"
)

// Handle the "verify" command.
func Verify(ctx context.Context, logger *Logger, arguments arguments) error {
	manifest, err := readManifest(logger, arguments.path)
	if errors.Is(err, fs.ErrNotExist) {
		return userErrorf(err, "No installation found at %s to verify.", arguments.path)
	}
	if err != nil {
		return userErrorf(err, "The installation's manifest at %s couldn't be read.", arguments.path)
	}

	// Repairing a damaged asset stages the download through the temp directory, so clear it on the way
	// out regardless of the return condition - including when the run is interrupted mid-download.
	defer clearTempDirectory(arguments.path)

	damaged := checkAssets(ctx, logger, arguments.path, manifest, "Damaged")
	// Assets that were never checked aren't known to be healthy, so a cancelled run
	// can't report a verdict on the installation.
	if err := ctx.Err(); err != nil {
		return err
	}

	if len(damaged) != 0 {
		message := logger.plural(len(damaged), "asset is damaged and needs", "assets are damaged and need") + " to be redownloaded. "
		logger.log(ASSETS_ARE_DAMAGED, fmt.Sprintf("%d %s", len(damaged), message))

		proceed := arguments.yes || logger.prompt(PROMPT_PROCEED)
		if proceed {
			logger.log(REDOWNLOADING)
			failed := downloadAssets(ctx, logger, arguments, damaged, "Downloaded")
			if err := ctx.Err(); err != nil {
				return err
			}
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
	verifyUtilities(ctx, logger, arguments)
	if err := ctx.Err(); err != nil {
		return err
	}

	logger.log(INSTALL_VERIFIED)
	return nil
}

// checkAssets verifies each asset in the manifest against what is on disk,
// logging per-asset progress, and returns the assets that need to be
// re-downloaded. `damagedLabel` is the word shown for a damaged asset
// (e.g. "Damaged" for verify, "Needs patching" for update).
func checkAssets(ctx context.Context, logger *Logger, root string, manifest Manifest, damagedLabel string) []Asset {
	counter := logger.count(len(manifest.Assets), "Verifying")
	defer counter.done()
	var mu sync.Mutex
	damaged := []Asset{}

	parallel(ctx, manifest.Assets, VERIFY_CONCURRENCY, func(asset Asset) {
		results := assets.VerifyAsset(ctx, root, asset)
		message := asset.Name
		if ctx.Err() != nil {
			message += ": " + logger.colour("Cancelled", GREY)
		} else if len(results) == 0 {
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
