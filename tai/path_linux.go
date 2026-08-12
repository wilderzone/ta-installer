package main

// samePath reports whether two cleaned paths refer to the same location.
// Linux filesystems are case-sensitive, so the comparison is exact.
func samePath(a string, b string) bool {
	return a == b
}
