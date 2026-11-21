package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSpaceNeededCountsArchivesAndUnpackedFiles(t *testing.T) {
	root := t.TempDir()
	toDownload := []Asset{
		{Name: "Core", Destination: "/Binaries", PackedSize: 100, UnpackedSize: 300, Files: []File{{File: "/game.exe"}}},
		{Name: "Maps", Destination: "/Maps", PackedSize: 50, UnpackedSize: 80, Files: []File{{File: "/map.upk"}}},
	}
	if got := spaceNeeded(root, toDownload); got != 530 {
		t.Errorf("got %d, want 530", got)
	}
}

func TestSpaceNeededCreditsTheFilesAlreadyThere(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "Binaries")
	if err := os.MkdirAll(directory, os.FileMode(0755)); err != nil {
		t.Fatalf("create directory: %v", err)
	}
	// A damaged copy, about to be replaced.
	if err := os.WriteFile(filepath.Join(directory, "game.exe"), make([]byte, 200), os.FileMode(0644)); err != nil {
		t.Fatalf("write file: %v", err)
	}

	asset := Asset{Name: "Core", Destination: "/Binaries", PackedSize: 100, UnpackedSize: 300, Files: []File{{File: "/game.exe"}, {File: "/missing.dll"}}}
	// The archive, plus the 100 bytes the new copy adds to the old.
	if got := spaceNeeded(root, []Asset{asset}); got != 200 {
		t.Errorf("got %d, want 200", got)
	}

	// Files already larger than the asset leave only the archive to make room for.
	asset.UnpackedSize = 150
	if got := spaceNeeded(root, []Asset{asset}); got != 100 {
		t.Errorf("got %d, want 100", got)
	}
}

func TestFreeSpaceMeasuresTheNearestExistingDirectory(t *testing.T) {
	root := t.TempDir()
	free, err := freeSpace(filepath.Join(root, "Not", "Created", "Yet"))
	if err != nil {
		t.Fatalf("measure free space: %v", err)
	}
	want, err := volumeFreeSpace(root)
	if err != nil {
		t.Fatalf("measure free space: %v", err)
	}
	// Other processes write to the same drive, so allow for a little drift.
	if difference := int64(free) - int64(want); difference > 1 << 20 || difference < -(1 << 20) {
		t.Errorf("got %d, want about %d", free, want)
	}
}
