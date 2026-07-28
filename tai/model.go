package main

import "wilderzone.org/tai/internal/assets"

// The CLI works in terms of the asset model defined in the assets package. These aliases
// let the rest of package main use the core types unqualified, while the behaviour over
// them lives in - and is unit-tested in - the assets package.
type (
	Asset      = assets.Asset
	File       = assets.File
	Manifest   = assets.Manifest
	FileResult = assets.FileResult
)
