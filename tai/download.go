package main

import (
	"archive/zip"
	"context"
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
func downloadAssets(ctx context.Context, logger *Logger, arguments arguments, assets []Asset, successLabel string) []Asset {
	counter := logger.count(len(assets), "Downloading")
	defer counter.done()
	var mu sync.Mutex
	failed := []Asset{}
	parallel(ctx, assets, DOWNLOAD_CONCURRENCY, func(asset Asset) {
		message := asset.Name
		err := download(ctx, arguments.repo, arguments.path, asset)
		if ctx.Err() != nil {
			// Whatever error a cancelled download reported is a consequence of the
			// cancellation, not a failure the user should be told to retry.
			message += ": " + logger.colour("Cancelled", GREY)
		} else if err != nil {
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

// get issues a cancellable GET and returns the response, whose body the caller must close.
// A non-OK status is an error, so callers only ever see a body worth reading.
func get(ctx context.Context, source string) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return nil, fmt.Errorf("build request for %s: %w", source, err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", source, err)
	}
	if response.StatusCode != http.StatusOK {
		response.Body.Close()
		return nil, fmt.Errorf("request %s: unexpected status %s", source, response.Status)
	}
	return response, nil
}

// clearTempDirectory removes the scratch space that downloads stage through. Each download
// tidies its own file away as it finishes, but that is best-effort - on Windows os.Remove
// fails outright while anything outside this process still holds the file open, which a
// virus scanner working through a freshly written asset will. Sweeping the directory once
// the command is finished with it is the backstop that keeps a missed file from staying in
// the installation for good. Removing a directory that was never created is not an error,
// so this is safe to defer from a command that may not download anything at all.
func clearTempDirectory(root string) {
	_ = os.RemoveAll(filepath.Join(root, TEMP_DIRECTORY))
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

func download(ctx context.Context, repo string, root string, asset Asset) error {
	// Download the asset into the temp directory.
	tempDirectory := filepath.Join(root, TEMP_DIRECTORY)
	err := os.MkdirAll(tempDirectory, os.FileMode(0777))
	if err != nil {
		return fmt.Errorf("create temp directory: %w", err)
	}
	file, err := os.CreateTemp(tempDirectory, "download-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	fileName := file.Name()
	// Discard the download on the way out, regardless of the return condition. A run that
	// is cancelled or fails part-way through would otherwise leave the partial file behind,
	// and nothing ever picks one back up - the next attempt starts a new one. Leaving the
	// handle open is worse still on Windows, where it keeps the temp directory itself from
	// being removed.
	defer func() {
		_ = file.Close() // A no-op once the successful path below has closed it.
		_ = os.Remove(fileName)
	}()

	source, err := url.JoinPath(repo, asset.Source)
	if err != nil {
		return fmt.Errorf("build source URL: %w", err)
	}
	response, err := get(ctx, source)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	// A cancelled context closes the response body, so this stops of its own accord.
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return fmt.Errorf("download %s: %w", source, err)
	}
	err = file.Close()
	if err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	// Unpack the downloaded file into the destination directory.
	err = unpack(ctx, root, fileName, asset)
	if err != nil {
		return fmt.Errorf("unpack %s: %w", asset.Name, err)
	}

	return nil
}

func unpack(ctx context.Context, root string, archivePath string, asset Asset) error {
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
		// Expanding a large archive takes a while, so give up between files rather
		// than only once the whole archive has been unpacked.
		if err := ctx.Err(); err != nil {
			return err
		}

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
