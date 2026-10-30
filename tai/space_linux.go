package main

import (
	"golang.org/x/sys/unix"
)

// volumeFreeSpace returns how many bytes can be written to the filesystem that the existing
// path is on. The blocks a filesystem reserves for root aren't available to the user, so
// they aren't counted.
func volumeFreeSpace(path string) (uint64, error) {
	stat := unix.Statfs_t{}
	err := unix.Statfs(path, &stat)
	if err != nil {
		return 0, err
	}
	return stat.Bavail * uint64(stat.Bsize), nil
}
