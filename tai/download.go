package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
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
	// Assets are only counted one at a time, and a large one takes minutes to arrive, so
	// the bytes are measured too. That is what moves while an asset is in flight.
	progress := counter.measure(totalPackedSize(assets))
	var mu sync.Mutex
	failed := []Asset{}
	parallel(ctx, assets, DOWNLOAD_CONCURRENCY, func(asset Asset) {
		message := asset.Name
		err := download(ctx, arguments.repo, arguments.path, asset, progress)
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

// totalPackedSize returns the number of bytes the given assets will transfer, or 0 if any
// one of them doesn't declare a size. A partial total would only mislead: the download
// would appear to overshoot it, and every estimate drawn from it would be wrong.
func totalPackedSize(assets []Asset) uint64 {
	total := uint64(0)
	for _, asset := range assets {
		if asset.PackedSize == 0 {
			return 0
		}
		total += asset.PackedSize
	}
	return total
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

// download fetches an asset and unpacks it into the installation. `progress`, when the
// caller is measuring a batch of them, counts the bytes as they arrive; a nil one is fine
// for a download nobody is watching that closely.
func download(ctx context.Context, repo string, root string, asset Asset, progress *Progress) error {
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
	_, err = io.Copy(file, progress.reader(response.Body))
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

// unpack expands a downloaded archive into the installation. The archive format is taken
// from the asset's source name rather than sniffed, so a mislabelled asset fails loudly
// instead of being unpacked as something it isn't.
func unpack(ctx context.Context, root string, archivePath string, asset Asset) error {
	destination := filepath.Join(root, asset.Destination)
	err := os.MkdirAll(destination, os.FileMode(0777))
	if err != nil {
		return fmt.Errorf("create destination directory: %w", err)
	}

	source := strings.ToLower(asset.Source)
	switch {
		case strings.HasSuffix(source, ".tar.gz"), strings.HasSuffix(source, ".tgz"):
			return unpackTar(ctx, destination, archivePath)

		default:
			return unpackZip(ctx, destination, archivePath)
	}
}

// resolveArchivePath joins an entry name onto the destination and rejects anything that
// would escape it, which is how an archive smuggles files into the rest of the filesystem.
func resolveArchivePath(destination string, name string) (string, error) {
	path := filepath.Join(destination, name)
	isWithinDestination := strings.HasPrefix(path, filepath.Clean(destination) + string(os.PathSeparator))
	if !isWithinDestination {
		return "", fmt.Errorf("illegal path in archive: %q", name)
	}
	return path, nil
}

// writeArchiveFile writes one entry's content. The mode comes from the archive: on Linux an
// executable that arrives without its execute bit is unusable, and both the CLI installed
// into the utilities directory and the compatibility runtime's binaries depend on it.
func writeArchiveFile(path string, mode os.FileMode, content io.Reader) error {
	err := os.MkdirAll(filepath.Dir(path), os.FileMode(0777))
	if err != nil {
		return err
	}
	// Truncating an existing file would fail with ETXTBSY if it happened to be running,
	// and would leave the old permissions in place.
	err = os.Remove(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	output, err := os.OpenFile(path, os.O_WRONLY | os.O_CREATE | os.O_EXCL, mode.Perm())
	if err != nil {
		return err
	}
	_, err = io.Copy(output, content)
	if closeErr := output.Close(); err == nil {
		err = closeErr
	}
	return err
}

func unpackZip(ctx context.Context, destination string, archivePath string) error {
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

		path, err := resolveArchivePath(destination, file.Name)
		if err != nil {
			return err
		}

		// Handle directories.
		if file.FileInfo().IsDir() {
			err = os.MkdirAll(path, os.FileMode(0777))
			if err != nil {
				return fmt.Errorf("create directory %s: %w", file.Name, err)
			}
			continue
		}

		archived, err := file.Open()
		if err != nil {
			return fmt.Errorf("read %s from archive: %w", file.Name, err)
		}
		// A zip written on Windows carries no Unix permissions at all, so fall back to
		// a readable, writable file rather than creating an unreadable one.
		mode := file.Mode().Perm()
		if mode == 0 {
			mode = os.FileMode(0666)
		}
		err = writeArchiveFile(path, mode, archived)
		archived.Close()
		if err != nil {
			return fmt.Errorf("write %s: %w", file.Name, err)
		}
	}

	return nil
}

// unpackTar expands a gzipped tar. The compatibility runtime ships as one: it carries
// symlinks and execute bits that a zip either can't represent or loses on the way through a
// Windows packaging step, and it keeps the published archive byte-identical to upstream's.
func unpackTar(ctx context.Context, destination string, archivePath string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer file.Close()
	decompressed, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer decompressed.Close()

	reader := tar.NewReader(decompressed)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("read archive: %w", err)
		}

		path, err := resolveArchivePath(destination, header.Name)
		if err != nil {
			return err
		}

		switch header.Typeflag {
			case tar.TypeDir:
				err = os.MkdirAll(path, os.FileMode(0777))
				if err != nil {
					return fmt.Errorf("create directory %s: %w", header.Name, err)
				}

			case tar.TypeReg:
				err = writeArchiveFile(path, os.FileMode(header.Mode), reader)
				if err != nil {
					return fmt.Errorf("write %s: %w", header.Name, err)
				}

			case tar.TypeSymlink:
				err = writeArchiveLink(destination, path, header.Linkname)
				if err != nil {
					return fmt.Errorf("link %s: %w", header.Name, err)
				}

			case tar.TypeLink:
				target, err := resolveArchivePath(destination, header.Linkname)
				if err != nil {
					return err
				}
				_ = os.Remove(path)
				err = os.Link(target, path)
				if err != nil {
					return fmt.Errorf("link %s: %w", header.Name, err)
				}
		}
	}

	return nil
}

// writeArchiveLink creates a symlink, rejecting one that points outside the destination.
// An absolute or upward-reaching target would let an archive overwrite files elsewhere on
// the system the next time anything wrote through the link.
func writeArchiveLink(destination string, path string, target string) error {
	if filepath.IsAbs(target) {
		return fmt.Errorf("illegal absolute link target: %q", target)
	}
	resolved := filepath.Join(filepath.Dir(path), target)
	if !strings.HasPrefix(resolved, filepath.Clean(destination) + string(os.PathSeparator)) {
		return fmt.Errorf("illegal link target: %q", target)
	}
	err := os.Remove(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	err = os.MkdirAll(filepath.Dir(path), os.FileMode(0777))
	if err != nil {
		return err
	}
	return os.Symlink(target, path)
}
