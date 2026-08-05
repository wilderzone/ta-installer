package main

// samePath reports whether two cleaned paths refer to the same location.
//
// The comparison is exact. Linux filesystems are case-sensitive, so folding case would make
// a sibling directory named "installer" indistinguishable from "Installer" - and the one
// caller uses this to decide whether it may delete that directory.
func samePath(a string, b string) bool {
	return a == b
}
