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
// The redistributable is still published for Linux and can be installed into the prefix if a
// build ever turns out to need the real d3dx9 helper libraries rather than the runtime's
// built-in ones. See installDirectX below.
func prerequisites(ctx context.Context, logger *Logger, arguments arguments) error {
	return ensureRuntime(ctx, logger, arguments.path)
}
