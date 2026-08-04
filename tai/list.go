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

// byteCountSI converts a given byte count to a string formatted with SI units.
func byteCountSI(bytes uint64) string {
	const unit = 1000
	const magnitudes = "kMGTPE"

	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	divisor, exponent := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		divisor *= unit
		exponent++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(divisor), magnitudes[exponent])
}

// tableHeadingRule produces a horizontal rule with a length matching the given string.
func tableHeadingRule(length string) string {
	var rule strings.Builder
	for range len(length) {
		rule.WriteString("-")
	}
	return rule.String()
}

// tableLeftPad pads the start of a given string with spaces until it reaches the specified length.
func tableLeftPad(s string, length int) string {
	var pad strings.Builder
	for range length - len(s) {
		pad.WriteString(" ")
	}
	pad.WriteString(s)
	return pad.String()
}

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
		packedSizeTotal := uint64(0)
		unpackedSizeTotal := uint64(0)
		for _, asset := range manifest.Assets {
			packedSizeTotal += asset.PackedSize
			unpackedSizeTotal += asset.UnpackedSize
		}
		packedSizeHeading := fmt.Sprintf("Packed (%s)", byteCountSI(packedSizeTotal))
		unpackedSizeHeading := fmt.Sprintf("Unpacked (%s)", byteCountSI(unpackedSizeTotal))
		fmt.Fprintf(w, "Name\tType\tFiles\t%s\t%s\tRequires\t\n", packedSizeHeading, unpackedSizeHeading)
		fmt.Fprintf(w, "----\t----\t-----\t%s\t%s\t--------\t\n", tableHeadingRule(packedSizeHeading), tableHeadingRule(unpackedSizeHeading))
		for _, asset := range manifest.Assets {
			packedSize := byteCountSI(asset.PackedSize)
			unpackedSize := byteCountSI(asset.UnpackedSize)
			fmt.Fprintf(w, "%s\t%s\t%5d\t%s\t%s\t%s\t\n",
				asset.Name,
				asset.Type,
				len(asset.Files),
				tableLeftPad(packedSize, len(packedSizeHeading)),
				tableLeftPad(unpackedSize, len(unpackedSizeHeading)),
				strings.Join(asset.Dependencies, ", "),
			)
		}
		w.Flush()
	}

	return nil
}
