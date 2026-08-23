package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// writeTestTar builds a gzipped tar in the shape the compatibility runtime arrives in: an
// executable, a plain file, and a relative symlink between them.
func writeTestTar(t *testing.T, path string) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create archive: %v", err)
	}
	defer file.Close()
	compressor := gzip.NewWriter(file)
	defer compressor.Close()
	archive := tar.NewWriter(compressor)
	defer archive.Close()

	write := func(header *tar.Header, body string) {
		if err := archive.WriteHeader(header); err != nil {
			t.Fatalf("write header: %v", err)
		}
		if body != "" {
			if _, err := archive.Write([]byte(body)); err != nil {
				t.Fatalf("write body: %v", err)
			}
		}
	}

	write(&tar.Header{Name: "bin", Typeflag: tar.TypeDir, Mode: 0755}, "")
	write(&tar.Header{Name: "bin/wine", Typeflag: tar.TypeReg, Mode: 0755, Size: 4}, "wine")
	write(&tar.Header{Name: "notes.txt", Typeflag: tar.TypeReg, Mode: 0644, Size: 5}, "hello")
	write(&tar.Header{Name: "bin/wine64", Typeflag: tar.TypeSymlink, Linkname: "wine", Mode: 0777}, "")
}

func TestUnpackTarPreservesModesAndLinks(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permissions and symlinks aren't meaningful here")
	}
	root := t.TempDir()
	archive := filepath.Join(root, "runtime.tar.gz")
	writeTestTar(t, archive)

	asset := Asset{Source: "/runtime/GE-Proton.tar.gz", Destination: "/Runtime"}
	err := unpack(context.Background(), root, archive, asset)
	if err != nil {
		t.Fatalf("unpack: %v", err)
	}

	// The runtime is useless if its binaries arrive without their execute bit.
	binary := filepath.Join(root, "Runtime", "bin", "wine")
	info, err := os.Stat(binary)
	if err != nil {
		t.Fatalf("stat unpacked binary: %v", err)
	}
	if info.Mode().Perm() != 0755 {
		t.Errorf("got mode %v, want 0755", info.Mode().Perm())
	}

	plain, err := os.Stat(filepath.Join(root, "Runtime", "notes.txt"))
	if err != nil {
		t.Fatalf("stat unpacked file: %v", err)
	}
	if plain.Mode().Perm() != 0644 {
		t.Errorf("got mode %v, want 0644", plain.Mode().Perm())
	}

	target, err := os.Readlink(filepath.Join(root, "Runtime", "bin", "wine64"))
	if err != nil {
		t.Fatalf("read unpacked symlink: %v", err)
	}
	if target != "wine" {
		t.Errorf("got link target %q, want \"wine\"", target)
	}
}

func TestUnpackTarRejectsEscapingPaths(t *testing.T) {
	root := t.TempDir()
	archive := filepath.Join(root, "evil.tar.gz")

	file, err := os.Create(archive)
	if err != nil {
		t.Fatalf("create archive: %v", err)
	}
	compressor := gzip.NewWriter(file)
	writer := tar.NewWriter(compressor)
	header := &tar.Header{Name: "../escaped.txt", Typeflag: tar.TypeReg, Mode: 0644, Size: 3}
	if err := writer.WriteHeader(header); err != nil {
		t.Fatalf("write header: %v", err)
	}
	if _, err := writer.Write([]byte("bad")); err != nil {
		t.Fatalf("write body: %v", err)
	}
	writer.Close()
	compressor.Close()
	file.Close()

	asset := Asset{Source: "/runtime/evil.tar.gz", Destination: "/Runtime"}
	if err := unpack(context.Background(), root, archive, asset); err == nil {
		t.Error("an entry pointing outside the destination must be rejected")
	}
}

func TestUnpackZipAppliesStoredModes(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permissions aren't meaningful here")
	}
	root := t.TempDir()
	archive := filepath.Join(root, "installer.zip")

	file, err := os.Create(archive)
	if err != nil {
		t.Fatalf("create archive: %v", err)
	}
	writer := zip.NewWriter(file)
	header := &zip.FileHeader{Name: "tai-cli", Method: zip.Deflate}
	header.SetMode(os.FileMode(0755))
	entry, err := writer.CreateHeader(header)
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}
	if _, err := entry.Write([]byte("cli")); err != nil {
		t.Fatalf("write entry: %v", err)
	}
	writer.Close()
	file.Close()

	asset := Asset{Source: "/installer/Installer-linux-amd64.zip", Destination: "/Installer"}
	if err := unpack(context.Background(), root, archive, asset); err != nil {
		t.Fatalf("unpack: %v", err)
	}

	// The desktop entries launch this binary, so it has to come out executable.
	info, err := os.Stat(filepath.Join(root, "Installer", "tai-cli"))
	if err != nil {
		t.Fatalf("stat unpacked binary: %v", err)
	}
	if info.Mode().Perm() != 0755 {
		t.Errorf("got mode %v, want 0755", info.Mode().Perm())
	}
}
