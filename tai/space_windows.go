package main

import (
	"golang.org/x/sys/windows"
)

// volumeFreeSpace returns how many bytes can be written to the drive that the existing path
// is on. Only what's available to this user is counted, so a disk quota is respected.
func volumeFreeSpace(path string) (uint64, error) {
	directory, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	available := uint64(0)
	err = windows.GetDiskFreeSpaceEx(directory, &available, nil, nil)
	if err != nil {
		return 0, err
	}
	return available, nil
}
