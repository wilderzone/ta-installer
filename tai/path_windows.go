package main

import "strings"

// samePath reports whether two cleaned paths refer to the same location.
// Windows paths aren't case-sensitive, so the comparison uses Unicode case-folding.
func samePath(a string, b string) bool {
	return strings.EqualFold(a, b)
}
