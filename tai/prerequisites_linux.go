package main

import (
	"context"
)

// prerequisites prepares everything the game needs beyond its own files.
//
// On Windows that means installing the DirectX 9 runtime into the system directory. Here it
// means the compatibility layer: the game's Direct3D 9 calls are served by the runtime's own
// implementation, so there is no DirectX package to install and - because everything TAI
// touches is under the user's home - nothing to elevate for either.
//
// If a build ever turns out to need the real d3dx9 helper libraries rather than the
// runtime's built-in ones, the redistributable the Windows path downloads can be run inside
// the prefix by the same two stages, without the elevation - and the overrides that make the
// game prefer them go in WINEDLLOVERRIDES at launch, rather than being written into the
// prefix, so they stay reversible.
func prerequisites(ctx context.Context, logger *Logger, arguments arguments) error {
	err := downloadRuntime(ctx, logger, arguments.path)
	if err != nil {
		return err
	}
	return ensureRuntime(ctx, logger, arguments.path)
}
