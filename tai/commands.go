package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

// Handle the "asset" command.
func GenerateAssetFile(arguments arguments) (string, error) {
	asset, err := CreateAsset(arguments.path, arguments.name)
	if err != nil {
		return "", errors.New("Failed to create asset file: " + err.Error())
	}
	path := filepath.Join(arguments.path, arguments.name+".json")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return "", errors.New("Failed to create asset file: " + err.Error())
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")
	err = encoder.Encode(asset)
	if err != nil {
		return "", errors.New("Failed to encode asset file: " + err.Error())
	}
	plural := "file:"
	if len(asset.Files) != 1 {
		plural = "files:"
	}
	message := fmt.Sprint("Created asset containing", len(asset.Files), plural, "\n", path)
	return message, nil
}

// Handle the "manifest" command.
func GenerateManifestFile(arguments arguments) (string, error) {
	manifest, err := CreateManifest(arguments.path)
	if err != nil {
		return "", errors.New("Failed to create manifest file: " + err.Error())
	}
	path := filepath.Join(arguments.path, "manifest.json")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		return "", errors.New("Failed to create manifest file: " + err.Error())
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")
	err = encoder.Encode(manifest)
	if err != nil {
		return "", errors.New("Failed to encode manifest file: " + err.Error())
	}
	plural := "asset:"
	if len(manifest.Assets) != 1 {
		plural = "assets:"
	}
	message := fmt.Sprint("Created manifest containing", len(manifest.Assets), plural, "\n", path)
	return message, nil
}

// Handle the "install" command.
func Install(arguments arguments) (string, error) {
	manifest, err := DownloadManifest(arguments.repo, arguments.path)
	if err != nil {
		return "", errors.New("Failed to download manifest: " + err.Error())
	}

	// Download and unpack assets.
	failed := []Asset{}
	prerequisitesFailed := false
	for index, asset := range manifest.Assets {
		message := fmt.Sprintf("\x1b[2m[%d/%d]\x1b[0m ", index + 1, len(manifest.Assets))
		message += asset.Name
		err = Download(arguments.repo, arguments.path, asset)
		if err != nil {
			message += ": \x1b[31mFailed\x1b[0m: " + err.Error()
			failed = append(failed, asset)
			if asset.Type == PREREQUISITE {
				prerequisitesFailed = true
			}
		} else {
			message += ": \x1b[32mDownloaded\x1b[0m"
		}
		fmt.Println(message)
	}

	// Install prerequisites.
	if !prerequisitesFailed {
		fmt.Println("Installing prerequisites...")
		err = Prerequisites(arguments.path)
		if err != nil {
			return "", errors.New("Failed to install prerequisites: " + err.Error())
		}
	}

	message := "Installed."
	if len(failed) == len(manifest.Assets) {
		message = "Installation failed."
	} else if len(failed) != 0 {
		message = "Partially installed. You may need to run the \"verify\" command."
	}
	return message, nil
}

// Handle the "update" command.
func Update(arguments arguments) (string, error) {
	fmt.Println("Checking for updates...")

	localManifest, err := ReadManifest(arguments.path)
	if err != nil {
		return "", errors.New("Failed to read local manifest: " + err.Error())
	}
	remoteManifest, err := GetManifest(arguments.repo)
	if err != nil {
		return "", errors.New("Failed to download manifest: " + err.Error())
	}

	// Compare the version fields of the local and remote manifests.
	if localManifest.Version >= remoteManifest.Version {
		return fmt.Sprintf("Installation is up to date: v%d", localManifest.Version), nil
	}
	fmt.Printf("Update found: v%d → v%d\n", localManifest.Version, remoteManifest.Version)

	// If the remote version is greater, save the new manifest, and then patch the local assets.
	manifest, err := StoreManifest(remoteManifest, arguments.path)
	if err != nil {
		return "", errors.New("Failed to save new manifest: " + err.Error())
	}

	fmt.Println("Checking assets...")
	patch, _ := CompareAssets(arguments.path, localManifest, manifest)
	if len(patch) == 0 {
		return "Updated.", nil
	}

	fmt.Println("Patching assets...")
	failed := []Asset{}
	for index, asset := range patch {
		message := fmt.Sprintf("\x1b[2m[%d/%d]\x1b[0m ", index + 1, len(patch))
		message += asset.Name
		err = Download(arguments.repo, arguments.path, asset)
		if err != nil {
			message += ": \x1b[31mFailed\x1b[0m: " + err.Error()
			failed = append(failed, asset)
		} else {
			message += ": \x1b[32mPatched\x1b[0m"
		}
		fmt.Println(message)
	}

	message := "Updated."
	if len(failed) == len(manifest.Assets) {
		message = "Update failed."
	} else if len(failed) != 0 {
		message = "Partially updated. You may need to run the \"verify\" command."
	}
	return message, nil
}

// Handle the "verify" command.
func Verify(arguments arguments) (string, error) {
	manifest, err := ReadManifest(arguments.path)
	if err != nil {
		return "", errors.New("Failed to read local manifest: " + err.Error())
	}

	damaged, _ := CheckForDamagedAssets(arguments.path, manifest)

	if len(damaged) != 0 {
		plural := "asset is damaged and needs"
		if len(damaged) != 1 {
			plural = "assets are damaged and need"
		}
		plural += " to be redownloaded. "
		fmt.Printf("%d %s", len(damaged), plural)

		proceed := arguments.yes
		if !proceed {
			fmt.Print("Proceed? [y/N] ")
			action := "n"
			_, _ = fmt.Scanln(&action)
			match, _ := regexp.MatchString("^[yY]", action)
			proceed = match
		}
		if proceed {
			fmt.Println("Redownloading...")
			failed := []Asset{}
			for index, asset := range damaged {
				message := fmt.Sprintf("\x1b[2m[%d/%d]\x1b[0m ", index + 1, len(damaged))
				message += asset.Name
				err = Download(arguments.repo, arguments.path, asset)
				if err != nil {
					message += ": \x1b[31mFailed\x1b[0m: " + err.Error()
					failed = append(failed, asset)
				} else {
					message += ": \x1b[32mDownloaded\x1b[0m"
				}
				fmt.Println(message)
			}

			if len(failed) != 0 {
				return "", errors.New("Failed to patch.")
			}
		}
	}

	message := "Installation verified."
	return message, nil
}
