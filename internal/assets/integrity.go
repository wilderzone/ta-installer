package assets

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// cancellable wraps a reader so that a long read stops promptly once ctx is
// cancelled, rather than running to completion first. Hashing a multi-gigabyte
// asset is a single io.Copy, which reads in 32KB chunks, so the check happens
// often enough to be responsive and is far too cheap to measure.
type cancellable struct {
	ctx    context.Context
	reader io.Reader
}

func (c cancellable) Read(p []byte) (int, error) {
	if err := c.ctx.Err(); err != nil {
		return 0, err
	}
	return c.reader.Read(p)
}

// checksumReader streams r through SHA256 and returns the hex-encoded digest,
// without buffering the whole input in memory.
func checksumReader(r io.Reader) (string, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// checksumFile returns the SHA256 digest of the file at path, streaming it
// rather than reading it entirely into memory. It abandons the file part-way
// through if ctx is cancelled, returning context.Canceled.
func checksumFile(ctx context.Context, path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	return checksumReader(cancellable{ctx: ctx, reader: file})
}

// FileState describes the outcome of verifying a single file against the manifest.
// A damaged file is a normal *result* of verification, not an error.
type FileState int

const (
	FileOK         FileState = iota // Present and matches its checksum.
	FileMissing                     // Absent from the installation.
	FileCorrupt                     // Present, but the checksum does not match.
	FileUnreadable                  // Present, but could not be read (e.g. permissions).
)

// FileResult is the verification finding for a single file.
type FileResult struct {
	File  File
	State FileState
	Err   error // Set only when State == FileUnreadable.
}

// VerifyAsset checks every file of an asset against the manifest and returns a
// finding for each file that is not intact. An empty result means the asset is
// healthy. This never returns an error: a missing or corrupt file is a finding.
//
// Cancelling ctx stops the check where it stands, so the findings are only
// complete while ctx is still live. Callers must check it before reading an
// empty result as a healthy asset.
func VerifyAsset(ctx context.Context, root string, asset Asset) []FileResult {
	results := []FileResult{}

	for _, entry := range asset.Files {
		// A file that was never opened has not been found wanting. Checking here as
		// well catches the files that never reach a read, such as missing ones.
		if ctx.Err() != nil {
			return results
		}
		path := filepath.Join(root, asset.Destination, entry.File)
		checksum, err := checksumFile(ctx, path)
		switch {
		case errors.Is(err, context.Canceled):
			// Cancelled part-way through: an unread file is not a damaged one.
			return results
		case errors.Is(err, fs.ErrNotExist):
			results = append(results, FileResult{File: entry, State: FileMissing})
		case err != nil:
			results = append(results, FileResult{File: entry, State: FileUnreadable, Err: err})
		case checksum != entry.Checksum:
			results = append(results, FileResult{File: entry, State: FileCorrupt})
		}
	}

	return results
}

// Describe renders a short parenthetical summary of file findings for the user,
// e.g. "(3 missing, 1 corrupt)".
func Describe(results []FileResult) string {
	var missing, corrupt, unreadable int
	for _, result := range results {
		switch result.State {
		case FileMissing:
			missing++
		case FileCorrupt:
			corrupt++
		case FileUnreadable:
			unreadable++
		}
	}

	parts := []string{}
	if missing > 0 {
		parts = append(parts, fmt.Sprintf("%d missing", missing))
	}
	if corrupt > 0 {
		parts = append(parts, fmt.Sprintf("%d corrupt", corrupt))
	}
	if unreadable > 0 {
		parts = append(parts, fmt.Sprintf("%d unreadable", unreadable))
	}
	return "(" + strings.Join(parts, ", ") + ")"
}
