package main

import (
	"io/fs"
	"path/filepath"
)

// directorySize computes the total size of a directory in bytes.
func directorySize(path string) (int64, error) {
	var size int64

	err := filepath.WalkDir(path, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Skip directories, since their entry size doesn't reflect content size.
		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				return err
			}
			size += info.Size()
		}
		return nil
	})

	return size, err
}
