// Package pathseg — Clean Architecture layer detection by the segments of a
// package's import path. The convention: a layer is defined by a sequence of
// segments, e.g. /dal/repository or /domain/model — regardless of the module
// prefix.
package pathseg

import "strings"

// Index returns the index of the first occurrence of seq as consecutive
// path segments, or -1.
func Index(path string, seq ...string) int {
	segs := Segments(path)
	if len(seq) == 0 || len(segs) < len(seq) {
		return -1
	}
	for i := 0; i+len(seq) <= len(segs); i++ {
		if matchAt(segs, i, seq) {
			return i
		}
	}
	return -1
}

// Contains reports whether the path contains seq as consecutive segments.
func Contains(path string, seq ...string) bool {
	return Index(path, seq...) >= 0
}

// EndsWith reports whether the path ends with the seq segments —
// i.e. the package is the root of the layer, not its subpackage.
func EndsWith(path string, seq ...string) bool {
	segs := Segments(path)
	if len(segs) < len(seq) {
		return false
	}
	return matchAt(segs, len(segs)-len(seq), seq)
}

// Segments splits an import path into segments.
func Segments(path string) []string {
	return strings.Split(path, "/")
}

func matchAt(segs []string, i int, seq []string) bool {
	for j, s := range seq {
		if segs[i+j] != s {
			return false
		}
	}
	return true
}
