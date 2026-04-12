package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Handle the "asset" command.
func GenerateAssetFile(logger *Logger, arguments arguments) error {
	asset, err := CreateAsset(arguments.path, arguments.name)
	if err != nil {
		return errors.New("Failed to create asset file: " + err.Error())
	}
	path := filepath.Join(arguments.path, arguments.name + ".json")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC, os.FileMode(0777))
	if err != nil {
		return errors.New("Failed to create asset file: " + err.Error())
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")
	err = encoder.Encode(asset)
	if err != nil {
		return errors.New("Failed to encode asset file: " + err.Error())
	}
	logger.log(ASSET_CREATED, "Created asset containing", len(asset.Files), logger.plural(len(asset.Files), "file: ", "files: "), path)
	return nil
}

// Handle the "manifest" command.
func GenerateManifestFile(logger *Logger, arguments arguments) error {
	manifest, err := CreateManifest(arguments.path)
	if err != nil {
		return errors.New("Failed to create manifest file: " + err.Error())
	}
	path := filepath.Join(arguments.path, "manifest.json")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC, os.FileMode(0777))
	if err != nil {
		return errors.New("Failed to create manifest file: " + err.Error())
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")
	err = encoder.Encode(manifest)
	if err != nil {
		return errors.New("Failed to encode manifest file: " + err.Error())
	}
	logger.log(MANIFEST_CREATED, "Created manifest containing", len(manifest.Assets), logger.plural(len(manifest.Assets), "asset: ", "assets: "), path)
	return nil
}

// Handle the "check" command.
func Check(logger *Logger, arguments arguments) error {
	_, err := ReadManifest(logger, arguments.path)
	if err != nil {
		logger.log(INSTALL_NOT_FOUND)
		return nil
	}

	logger.log(INSTALL_FOUND)
	return nil
}

// Handle the "install" command.
func Install(logger *Logger, arguments arguments) error {
	manifest, err := DownloadManifest(logger, arguments.repo, arguments.path)
	if err != nil {
		return errors.New("Failed to download manifest: " + err.Error())
	}

	// Download and unpack assets.
	counter := logger.count(len(manifest.Assets))
	failed := []Asset{}
	prerequisitesFailed := false
	for _, asset := range manifest.Assets {
		message := asset.Name
		err = Download(arguments.repo, arguments.path, asset)
		if err != nil {
			message += ": " + logger.colour("Failed", RED) + ": " + err.Error()
			failed = append(failed, asset)
			if asset.Type == PREREQUISITE {
				prerequisitesFailed = true
			}
		} else {
			message += ": " + logger.colour("Downloaded", GREEN)
		}
		counter.next(message)
	}

	// Install prerequisites.
	if !prerequisitesFailed {
		logger.log(INSTALLING_PREREQUISITES)
		err = Prerequisites(arguments.path)
		if err != nil {
			return errors.New("Failed to install prerequisites: " + err.Error())
		}
	}

	if len(failed) == len(manifest.Assets) {
		logger.log(INSTALL_FAILURE)
	} else if len(failed) != 0 {
		logger.log(INSTALL_PARTIAL)
	} else {
		logger.log(INSTALL_SUCCESS)
	}
	return nil
}

// Handle the "update" command.
func Update(logger *Logger, arguments arguments) error {
	logger.log(CHECKING_FOR_UPDATES)

	localManifest, err := ReadManifest(logger, arguments.path)
	if err != nil {
		return errors.New("Failed to read local manifest: " + err.Error())
	}
	remoteManifest, err := GetManifest(logger, arguments.repo)
	if err != nil {
		return errors.New("Failed to download manifest: " + err.Error())
	}

	// Compare the version fields of the local and remote manifests.
	if localManifest.Version >= remoteManifest.Version {
		logger.log(INSTALL_UP_TO_DATE, fmt.Sprintf("Installation is up to date: v%d", localManifest.Version))
		return nil
	}
	logger.log(UPDATE_FOUND, fmt.Sprintf("v%d → v%d\n", localManifest.Version, remoteManifest.Version))

	// If the remote version is greater, save the new manifest, and then patch the local assets.
	manifest, err := StoreManifest(remoteManifest, arguments.path)
	if err != nil {
		return errors.New("Failed to save new manifest: " + err.Error())
	}

	logger.log(CHECKING_ASSETS)
	patch, _ := CompareAssets(logger, arguments.path, localManifest, manifest)
	if len(patch) == 0 {
		logger.log(UPDATE_SUCCESS)
		return nil
	}

	logger.log(PATCHING_ASSETS)
	counter := logger.count(len(patch))
	failed := []Asset{}
	for _, asset := range patch {
		message := asset.Name
		err = Download(arguments.repo, arguments.path, asset)
		if err != nil {
			message += ": " + logger.colour("Failed", RED) + ": " + err.Error()
			failed = append(failed, asset)
		} else {
			message += ": " + logger.colour("Patched", GREEN)
		}
		counter.next(message)
	}

	if len(failed) == len(manifest.Assets) {
		logger.log(UPDATE_FAILURE)
	} else if len(failed) != 0 {
		logger.log(UPDATE_PARTIAL)
	} else {
		logger.log(UPDATE_SUCCESS)
	}
	return nil
}

// Handle the "verify" command.
func Verify(logger *Logger, arguments arguments) error {
	manifest, err := ReadManifest(logger, arguments.path)
	if err != nil {
		return errors.New("Failed to read local manifest: " + err.Error())
	}

	damaged, _ := CheckForDamagedAssets(logger, arguments.path, manifest)

	if len(damaged) != 0 {
		message := logger.plural(len(damaged), "asset is damaged and needs", "assets are damaged and need") + " to be redownloaded. "
		logger.log(ASSETS_ARE_DAMAGED, fmt.Sprintf("%d %s", len(damaged), message))

		proceed := arguments.yes || logger.prompt(PROMPT_PROCEED)
		if proceed {
			logger.log(REDOWNLOADING)
			counter := logger.count(len(damaged))
			failed := []Asset{}
			for _, asset := range damaged {
				message := asset.Name
				err = Download(arguments.repo, arguments.path, asset)
				if err != nil {
					message += ": " + logger.colour("Failed", RED) + ": " + err.Error()
					failed = append(failed, asset)
				} else {
					message += ": " + logger.colour("Downloaded", GREEN)
				}
				counter.next(message)
			}

			if len(failed) != 0 {
				return errors.New("Failed to patch.")
			}
		}
	}

	logger.log(INSTALL_VERIFIED)
	return nil
}
