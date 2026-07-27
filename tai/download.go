package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// downloadAssets downloads and unpacks each asset concurrently, logging per-asset
// progress with the given success label, and returns the assets that failed.
func downloadAssets(logger *Logger, arguments arguments, assets []Asset, successLabel string) []Asset {
	counter := logger.count(len(assets), "Downloading")
	defer counter.done()
	var mu sync.Mutex
	failed := []Asset{}
	parallel(assets, DOWNLOAD_CONCURRENCY, func(asset Asset) {
		message := asset.Name
		err := Download(arguments.repo, arguments.path, asset)
		if err != nil {
			message += ": " + logger.colour("Failed", RED) + ": " + err.Error()
			mu.Lock()
			failed = append(failed, asset)
			mu.Unlock()
		} else {
			message += ": " + logger.colour(successLabel, GREEN)
		}
		counter.next(message)
	})
	return failed
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true // File exists.
	}
	if errors.Is(err, os.ErrNotExist) {
		return false // File explicitly does not exist.
	}
	// File may or may not exist (e.g., permission denied, disk failure).
	return false 
}

func Download(repo string, root string, asset Asset) error {
	// Download the asset into the temp directory.
	tempDirectory := filepath.Join(root, TEMP_DIRECTORY)
	err := os.MkdirAll(tempDirectory, os.FileMode(0777))
	if err != nil {
		return fmt.Errorf("create temp directory: %w", err)
	}
	file, err := os.CreateTemp(tempDirectory, "download-*")
	fileName := file.Name()
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	source, err := url.JoinPath(repo, asset.Source)
	if err != nil {
		return fmt.Errorf("build source URL: %w", err)
	}
	response, err := http.Get(source)
	if err != nil {
		return fmt.Errorf("request %s: %w", source, err)
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("request %s: unexpected status %s", source, response.Status)
	}
	defer response.Body.Close()
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return fmt.Errorf("download %s: %w", source, err)
	}
	err = file.Close()
	if err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	// Unpack the downloaded file into the destination directory.
	err = Unpack(root, fileName, asset)
	if err != nil {
		return fmt.Errorf("unpack %s: %w", asset.Name, err)
	}
	err = os.Remove(fileName)
	if err != nil {
		return fmt.Errorf("remove temp file: %w", err)
	}

	return nil
}

func Unpack(root string, archivePath string, asset Asset) error {
	destination := filepath.Join(root, asset.Destination)
	err := os.MkdirAll(destination, os.FileMode(0777))
	if err != nil {
		return fmt.Errorf("create destination directory: %w", err)
	}
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		// Prevent zip-slips.
		path := filepath.Join(destination, file.Name)
		isWithinDestination := strings.HasPrefix(path, filepath.Clean(destination) + string(os.PathSeparator))
		if !isWithinDestination {
			return fmt.Errorf("illegal path in archive: %q", file.Name)
		}

		// Handle directories.
		if file.FileInfo().IsDir() {
			err = os.MkdirAll(path, os.FileMode(0777))
			if err != nil {
				return fmt.Errorf("create directory %s: %w", file.Name, err)
			}
			continue
		}

		// Open the destination file.
		output, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("create %s: %w", file.Name, err)
		}

		// Open the archived file.
		archived, err := file.Open()
		if err != nil {
			output.Close()
			return fmt.Errorf("read %s from archive: %w", file.Name, err)
		}

		// Copy the file content.
		_, err = io.Copy(output, archived)
		output.Close()
		archived.Close()

		if err != nil {
			return fmt.Errorf("write %s: %w", file.Name, err)
		}
	}

	return nil
}
