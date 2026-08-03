package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"text/tabwriter"

	"wilderzone.org/tai/internal/assets"
)

// Handle the "list" command.
func List(ctx context.Context, logger *Logger, arguments arguments) error {
	if logger.silent {
		return nil
	}

	var manifest Manifest
	var err error

	if len(arguments.path) != 0 {
		// List manifest assets in the specified directory.
		manifest, err = readManifest(logger, arguments.path)
		if errors.Is(err, fs.ErrNotExist) {
			logger.log(INSTALL_NOT_FOUND)
			return nil
		}
		if err != nil {
			return userErrorf(err, "The installation at %s has a manifest that couldn't be read.", arguments.path)
		}
	} else if len(arguments.repo) != 0 {
		// List manifest assets in the specified repo.
		manifest, err = getManifest(ctx, logger, arguments.repo)
		if err != nil {
			// An abandoned request says nothing about whether the repository is actually there.
			if err := ctx.Err(); err != nil {
				return err
			}
			logger.log(REPO_NOT_FOUND)
			return nil
		}
	}

	// Filter the manifest's assets based on the provided include/exclude arguments.
	manifest.Assets = assets.FilterAssets(manifest.Assets, arguments.include, arguments.exclude)

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
		fmt.Fprint(w, "Name\tType\tFiles\tRequires\t\n")
		fmt.Fprint(w, "----\t----\t-----\t--------\t\n")
		for _, asset := range manifest.Assets {
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\t\n", asset.Name, asset.Type, len(asset.Files), strings.Join(asset.Dependencies, ", "))
		}
		w.Flush()
	}

	return nil
}
