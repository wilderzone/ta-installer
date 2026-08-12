package main

import "context"

// These are the hooks that only mean something on a platform that runs the game through a
// compatibility layer. Windows runs it directly, so both are no-ops - which is what lets
// install.go and verify.go call them without knowing which platform they're on.

// preflight reports anything about this machine that would stop the game running, before a
// multi-gigabyte download commits the user to it. The DirectX 9 runtime is Windows' only
// requirement, and installing it is already part of the install rather than something the
// user has to arrange.
func preflight(logger *Logger) {}

// verifyRuntime checks the compatibility layer against its published checksums.
func verifyRuntime(ctx context.Context, logger *Logger, arguments arguments) {}
