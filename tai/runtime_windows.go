package main

import "context"

// These hooks are only meaningful on a platform that runs the game through a compatibility
// layer. Windows runs the game directly, so both hooks are empty.

// preflight reports anything about this machine that would stop the game running, before
// TAI commits to a multi-gigabyte download.
//
// Returns true if the preflight check was successful.
func preflight(logger *Logger) bool {
	return true
}

// verifyRuntime checks the compatibility layer against its published checksums.
func verifyRuntime(ctx context.Context, logger *Logger, arguments arguments) {}
