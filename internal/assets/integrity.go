package assets

import (
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
// rather than reading it entirely into memory.
func checksumFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	return checksumReader(file)
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
func VerifyAsset(root string, asset Asset) []FileResult {
	results := []FileResult{}

	for _, entry := range asset.Files {
		path := filepath.Join(root, asset.Destination, entry.File)
		checksum, err := checksumFile(path)
		switch {
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
