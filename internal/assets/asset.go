// Package assets defines the asset and manifest data model, and the pure,
// platform-independent logic over it: dependency resolution, filtering and merging, and
// integrity checking. Keeping it free of the CLI's I/O, logging, and Windows-specific code
// lets it build and be unit-tested on any platform.
package assets

type AssetType string

const (
	CORE         AssetType = "core"
	DLC          AssetType = "dlc"
	PREREQUISITE AssetType = "prerequisite"
)

type Asset struct {
	// The unique name of the asset.
	Name         string    `json:"name"`
	// The asset type.
	Type         AssetType `json:"type"`
	// The location of the asset's remote source.
	Source       string    `json:"source"`
	// The location of the asset's local destination.
	Destination  string    `json:"destination"`
	// The names of any other assets that this asset depends on.
	Dependencies []string  `json:"dependencies"`
	// Whether this asset was installed to satisfy another asset's dependency, rather
	// than being asked for directly. Only ever set in a local manifest.
	Automatic    bool      `json:"automatic,omitempty"`
	// The files contained within the asset.
	Files        []File    `json:"files"`
}

type File struct {
	// The file's checksum.
	Checksum string `json:"checksum"`
	// The file name (including relative path & extension).
	File     string `json:"file"`
}

type Manifest struct {
	Assets  []Asset `json:"assets"`
	Version uint32  `json:"version"`
}
