package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
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
func CheckForDamagedAssets(root string, manifest Manifest) ([]Asset, []error) {
	damaged := []Asset{}
	errs := []error{}

	for index, asset := range manifest.Assets {
		message := fmt.Sprintf("\x1b[2m[%d/%d]\x1b[0m ", index + 1, len(manifest.Assets))
		message += asset.Name
		missing, errs := CheckForMissingFiles(root, asset)
		// Mark as damaged if any files are missing.
		if len(missing) > 0 || len(errs) > 0 {
			// errs = append(errs, err)
			damaged = append(damaged, asset)
			message += ": \x1b[31mDamaged\x1b[0m"
		} else {
			message += ": \x1b[32mVerified\x1b[0m"
		}
		fmt.Println(message)
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
