package main

// preflight reports anything about this machine that would stop the game running, before a
// multi-gigabyte download commits the user to it. Windows runs the game directly, so there
// is nothing to check: the DirectX 9 runtime is the only requirement, and installing it is
// already part of the install rather than something the user has to arrange.
func preflight(logger *Logger) {}
