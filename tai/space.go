package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// spaceNeeded estimates the most disk space that downloading the given assets into the
// installation could take: every archive, plus whatever their unpacked files add beyond the
// copies already there.
//
// Archives are downloaded while others are still being unpacked, and each is only deleted
// once it has been, so all of them are counted - the estimate errs on the side of too much.
// Unpacking deletes each existing copy of a file before writing the new one, so the space
// an asset already takes up is given back as it's replaced.
func spaceNeeded(root string, toDownload []Asset) uint64 {
	needed := uint64(0)
	for _, asset := range toDownload {
		needed += asset.PackedSize
		existing := existingSize(root, asset)
		if asset.UnpackedSize > existing {
			needed += asset.UnpackedSize - existing
		}
	}
	return needed
}

// existingSize returns the space taken up by the copies of the asset's files that are
// already in the installation, damaged or not.
func existingSize(root string, asset Asset) uint64 {
	total := uint64(0)
	for _, file := range asset.Files {
		info, err := os.Stat(filepath.Join(root, asset.Destination, file.File))
		if err == nil && info.Mode().IsRegular() {
			total += uint64(info.Size())
		}
	}
	return total
}

// freeSpace returns how many bytes can be written to the drive that the path is on. The
// path needn't exist yet - an installation directory usually doesn't before it's installed -
// so the nearest directory above it that does is measured instead.
func freeSpace(path string) (uint64, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return 0, err
	}
	for {
		_, err := os.Stat(path)
		if err == nil {
			return volumeFreeSpace(path)
		}
		parent := filepath.Dir(path)
		if parent == path {
			return 0, err
		}
		path = parent
	}
}

// checkFreeSpace warns when the installation's drive may not have room for the given assets,
// and reports whether it did, so the caller can ask whether to go ahead anyway. The estimate
// is generous, so it never stops anything by itself, and free space that can't be measured is
// no reason to warn.
//
// The warning names the `activity` (eg, "installation") that needs the space, since the
// archives make it more than the installed size an interface may have shown beforehand.
func checkFreeSpace(logger *Logger, root string, toDownload []Asset, activity string) bool {
	needed := spaceNeeded(root, toDownload)
	if needed == 0 {
		return false
	}
	free, err := freeSpace(root)
	if err != nil {
		logger.detail("measure free space", err)
		return false
	}
	if free >= needed {
		return false
	}

	// The figures are carried as well, for an interface to word the warning its own way.
	line := INSUFFICIENT_SPACE
	line.Bytes = free
	line.Size = needed
	logger.log(line, fmt.Sprintf("%s is free at %s, and %s may be required during %s.", byteCountSI(free), root, byteCountSI(needed), activity))
	return true
}
