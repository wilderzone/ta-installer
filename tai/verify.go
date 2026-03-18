package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
)

// Calculate the SHA256 checksum for a given file.
func Checksum(file []byte) string {
	hash := sha256.New()
	hash.Write(file)
	bytes := hash.Sum(nil)
	return hex.EncodeToString(bytes)
}

// Given a manifest, check for any assets that have not been installed correctly.
func CheckForDamagedAssets(logger *Logger, root string, manifest Manifest) ([]Asset, []error) {
	counter := logger.count(len(manifest.Assets))
	damaged := []Asset{}
	errs := []error{}

	for _, asset := range manifest.Assets {
		message := asset.Name
		missing, errs := CheckForMissingFiles(root, asset)
		// Mark as damaged if any files are missing.
		if len(missing) > 0 || len(errs) > 0 {
			// errs = append(errs, err)
			damaged = append(damaged, asset)
			message += ": " + logger.colour("Damaged", RED)
		} else {
			message += ": " + logger.colour("Verified", GREEN)
		}
		counter.next(message)
	}

	return damaged, errs
}

// Given a manifest asset, check for any files missing from the installation.
func CheckForMissingFiles(root string, asset Asset) ([]File, []error) {
	missing := []File{}
	errs := []error{}

	for _, file := range asset.Files {
		path := filepath.Join(root, asset.Destination, file.File)
		content, err := os.ReadFile(path)
		// Mark as missing if the OS fails to read the file.
		if err != nil {
			errs = append(errs, err)
			missing = append(missing, file)
			continue
		}
		// Mark as missing if the checksum differs.
		checksum := Checksum(content)
		if checksum != file.Checksum {
			errs = append(errs, errors.New("Checksum mismatch."))
			missing = append(missing, file)
		}
	}

	return missing, errs
}

// Compare the assets listed in `currentManifest` to those in `newManifest`, returning a list of assets that need to be patched.
func CompareAssets(logger *Logger, root string, currentManifest Manifest, newManifest Manifest) ([]Asset, []error) {
	counter := logger.count(len(newManifest.Assets))
	patch := []Asset{}
	errs := []error{}

	for _, asset := range newManifest.Assets {
		message := asset.Name
		missing, errs := CheckForMissingFiles(root, asset)
		// Mark as needs-patching if any files are missing.
		if len(missing) > 0 || len(errs) > 0 {
			// errs = append(errs, err)
			patch = append(patch, asset)
			message += ": " + logger.colour("Needs patching", RED)
		} else {
			message += ": " + logger.colour("Verified", GREEN)
		}
		counter.next(message)
	}

	return patch, errs
}
