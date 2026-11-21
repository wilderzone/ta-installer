package main

import (
	"os"
	"path/filepath"
	"testing"

	"wilderzone.org/tai-utilities/internal/frontend"
)

// installation lays out an installation directory holding the utilities directory and the
// given entries, and returns where it is.
func installation(t *testing.T, entries ...string) frontend.Installation {
	game := t.TempDir()
	utilities := filepath.Join(game, "Installer")
	err := os.Mkdir(utilities, os.FileMode(0755))
	if err != nil {
		t.Fatalf("create the utilities directory: %v", err)
	}
	for _, entry := range entries {
		err := os.WriteFile(filepath.Join(game, entry), []byte{}, os.FileMode(0644))
		if err != nil {
			t.Fatalf("create %s: %v", entry, err)
		}
	}
	return frontend.Installation{Game: game, CLI: filepath.Join(utilities, "tai-cli")}
}

func TestOnlyUtilitiesRemainAfterACompleteUninstall(t *testing.T) {
	if !onlyUtilitiesRemain(installation(t)) {
		t.Error("got false for an installation holding only its utilities")
	}
}

func TestOnlyUtilitiesRemainSparesAnythingElse(t *testing.T) {
	// Whatever the CLI left behind, the installation directory must not be swept away with it.
	if onlyUtilitiesRemain(installation(t, "manifest.json")) {
		t.Error("got true for an installation the CLI left files in")
	}
}

func TestOnlyUtilitiesRemainSparesAMissingInstallation(t *testing.T) {
	found := frontend.Installation{Game: filepath.Join(t.TempDir(), "missing"), CLI: "tai-cli"}
	if onlyUtilitiesRemain(found) {
		t.Error("got true for an installation directory that couldn't be read")
	}
}
