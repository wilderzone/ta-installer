package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeELF writes a file with just enough of an ELF header to be identified.
func writeELF(t *testing.T, path string, class byte) {
	t.Helper()
	err := os.WriteFile(path, []byte{0x7f, 'E', 'L', 'F', class, 0, 0, 0}, os.FileMode(0644))
	if err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestIs32BitReadsTheELFClass(t *testing.T) {
	directory := t.TempDir()

	thirtyTwo := filepath.Join(directory, "libvulkan_radeon.so")
	sixtyFour := filepath.Join(directory, "libvulkan_radeon64.so")
	text := filepath.Join(directory, "not-a-library.so")
	writeELF(t, thirtyTwo, 1)
	writeELF(t, sixtyFour, 2)
	os.WriteFile(text, []byte("#!/bin/sh\n"), os.FileMode(0644))

	if !is32Bit(thirtyTwo) {
		t.Error("a 32-bit library wasn't recognised")
	}
	if is32Bit(sixtyFour) {
		t.Error("a 64-bit library was reported as 32-bit")
	}
	if is32Bit(text) {
		t.Error("a file that isn't ELF at all was reported as 32-bit")
	}
	if is32Bit(filepath.Join(directory, "absent.so")) {
		t.Error("a missing library was reported as 32-bit")
	}
}

func TestIs32BitLibraryResolvesRelativeToTheManifest(t *testing.T) {
	// The Vulkan loader resolves a library with a separator in it against the directory the
	// manifest sits in, which is how a driver installed outside the system paths is found.
	directory := t.TempDir()
	err := os.MkdirAll(filepath.Join(directory, "lib"), os.FileMode(0755))
	if err != nil {
		t.Fatalf("create the library directory: %v", err)
	}
	writeELF(t, filepath.Join(directory, "lib", "libvulkan_driver.so"), 1)

	manifest := filepath.Join(directory, "driver_icd.i686.json")
	if !is32BitLibrary(manifest, "./lib/libvulkan_driver.so") {
		t.Error("a library beside the manifest wasn't found")
	}
	if is32BitLibrary(manifest, "./lib/libvulkan_missing.so") {
		t.Error("a library that isn't there was reported as 32-bit")
	}
}

func TestInspectSeparatesSoftwareFromHardware(t *testing.T) {
	directory := t.TempDir()
	writeELF(t, filepath.Join(directory, "libvulkan_lvp.so"), 1)
	writeELF(t, filepath.Join(directory, "libvulkan_radeon.so"), 1)

	manifest := func(name string, library string) string {
		path := filepath.Join(directory, name)
		contents := `{"file_format_version":"1.0.0","ICD":{"library_path":"` + library + `"}}`
		if err := os.WriteFile(path, []byte(contents), os.FileMode(0644)); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
		return path
	}

	// A machine with only lavapipe registered has a 32-bit driver and still can't play the
	// game, which is exactly the case the old check couldn't see.
	software := vulkanDrivers{}
	software.inspect(manifest("lvp_icd.i686.json", "./libvulkan_lvp.so"))
	if !software.present || !software.present32 {
		t.Error("lavapipe wasn't counted as an installed 32-bit driver")
	}
	if software.hardware32 {
		t.Error("lavapipe was counted as hardware")
	}

	real := vulkanDrivers{}
	real.inspect(manifest("radeon_icd.i686.json", "./libvulkan_radeon.so"))
	if !real.hardware32 {
		t.Error("a real 32-bit driver wasn't counted as hardware")
	}
}

func TestDistributionFollowsDerivatives(t *testing.T) {
	cases := map[string]string{
		"ID=ubuntu\nID_LIKE=debian\n":                    "debian",
		"ID=linuxmint\nID_LIKE=\"ubuntu debian\"\n":      "debian",
		"ID=nobara\nID_LIKE=\"fedora\"\n":                "fedora",
		"ID=endeavouros\nID_LIKE=arch\n":                 "arch",
		"NAME=\"Some Distribution\"\nID=somethingelse\n": "",
		"":                                               "",
	}
	for contents, want := range cases {
		if got := distributionFrom(contents); got != want {
			t.Errorf("got %q, want %q, for %q", got, want, contents)
		}
	}
}

func TestDriverInstructionsNameTheRightPackages(t *testing.T) {
	mesa := vulkanDrivers{}
	command := driverInstructions(mesa, "debian")
	if !strings.Contains(command, "mesa-vulkan-drivers:i386") || !strings.Contains(command, "--add-architecture i386") {
		t.Errorf("the Debian command is missing the packages or the architecture: %q", command)
	}

	if command := driverInstructions(mesa, "fedora"); !strings.Contains(command, "mesa-vulkan-drivers.i686") {
		t.Errorf("the Fedora command is missing the packages: %q", command)
	}

	// An unrecognised distribution gets no command rather than a wrong one; the problem is
	// still reported on its own.
	if command := driverInstructions(mesa, ""); command != "" {
		t.Errorf("got %q for an unknown distribution, want no command", command)
	}
}
