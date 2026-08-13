package main

import "context"

// These hooks are only meaningful on a platform that runs the game through a compatibility
// layer. Windows runs the game directly, so both hooks are empty.

// preflight reports anything about this machine that would stop the game running, before
// committing to a multi-gigabyte download.
func preflight(logger *Logger) {}

// verifyRuntime checks the compatibility layer against its published checksums.
func verifyRuntime(ctx context.Context, logger *Logger, arguments arguments) {}
