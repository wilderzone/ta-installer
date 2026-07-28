package assets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestChecksumReaderKnownVector(t *testing.T) {
	// SHA-256 of "abc" is a well-known test vector.
	const want = "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"
	got, err := checksumReader(strings.NewReader("abc"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("checksum = %q, want %q", got, want)
	}
}

func TestDescribeCounts(t *testing.T) {
	results := []FileResult{
		{State: FileMissing},
		{State: FileMissing},
		{State: FileCorrupt},
	}
	if got := Describe(results); got != "(2 missing, 1 corrupt)" {
		t.Errorf("Describe = %q, want %q", got, "(2 missing, 1 corrupt)")
	}
}

func TestVerifyAssetDetectsCorruptAndMissing(t *testing.T) {
	root := t.TempDir()
	// A file that will match, and a file that will be corrupt on disk.
	writeFile(t, filepath.Join(root, "good.txt"), "hello")
	writeFile(t, filepath.Join(root, "bad.txt"), "actual contents")

	goodSum, err := checksumReader(strings.NewReader("hello"))
	if err != nil {
		t.Fatal(err)
	}

	asset := Asset{
		Destination: "",
		Files: []File{
			{File: "good.txt", Checksum: goodSum},
			{File: "bad.txt", Checksum: "0000"},  // wrong checksum -> corrupt
			{File: "gone.txt", Checksum: "abcd"}, // not on disk -> missing
		},
	}

	results := VerifyAsset(root, asset)
	if len(results) != 2 {
		t.Fatalf("got %d findings, want 2 (good.txt should pass): %+v", len(results), results)
	}
	states := map[string]FileState{}
	for _, r := range results {
		states[r.File.File] = r.State
	}
	if states["bad.txt"] != FileCorrupt {
		t.Errorf("bad.txt state = %v, want FileCorrupt", states["bad.txt"])
	}
	if states["gone.txt"] != FileMissing {
		t.Errorf("gone.txt state = %v, want FileMissing", states["gone.txt"])
	}
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
