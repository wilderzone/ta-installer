package main

// removeSettings does nothing on Linux. The game keeps its settings and logs in the Documents
// folder of its prefix, which is inside the installation, so they're removed along with it.
func removeSettings(logger *Logger) {}
