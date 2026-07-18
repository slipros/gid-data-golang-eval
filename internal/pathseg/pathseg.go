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

// HasLayer reports whether path belongs to the Clean Architecture layer
// identified by seq — i.e. seq matches the leading segments of the package's
// layer path (LayerSegments). Unlike Contains, the layer is anchored to the
// module root: a segment nested below another layer (e.g. a server-side
// package .../connect/client/interceptor) is NOT that layer, so use HasLayer
// (not Contains) whenever a rule classifies a package's own layer.
func HasLayer(path string, seq ...string) bool {
	segs := LayerSegments(path)
	if len(seq) == 0 || len(segs) < len(seq) {
		return false
	}
	for i, s := range seq {
		if segs[i] != s {
			return false
		}
	}
	return true
}

// LayerSegments returns the path segments after the module root — the layer
// path used to classify a package's layer. The module boundary is resolved in
// priority order: the /internal/ segment (canonical layout), then a
// /pkg/<module>/ segment (application-module layout — module.md), then
// (non-standard layout, e.g. testdata) the first path segment as the module
// root.
func LayerSegments(path string) []string {
	const internalSeg = "/internal/"
	if _, rest, ok := strings.Cut(path, internalSeg); ok {
		return nonEmpty(Segments(rest))
	}
	if root, ok := PkgModuleRoot(path); ok {
		rest := strings.TrimPrefix(strings.TrimPrefix(path, root), "/")
		return nonEmpty(Segments(rest))
	}
	_, rest, _ := strings.Cut(path, "/")
	return nonEmpty(Segments(rest))
}

// ModuleRoot returns the module-root prefix of a package path — the boundary
// used to tell whether two packages belong to the same module
// (ModuleRoot(a) == ModuleRoot(b)). Resolved in the same priority order as
// LayerSegments: the prefix before /internal/ (canonical layout), then a
// /pkg/<module> root (application-module layout — module.md), then the first
// path segment (non-standard layout). Note: comparing Segments(path)[0]
// directly is wrong for real import paths — e.g. every github.com/<org>/<repo>
// package shares the segment "github.com".
func ModuleRoot(path string) string {
	const internalSeg = "/internal/"
	if prefix, _, ok := strings.Cut(path, internalSeg); ok {
		return prefix
	}
	if root, ok := PkgModuleRoot(path); ok {
		return root
	}
	first, _, _ := strings.Cut(path, "/")
	return first
}

// PkgModuleRoot returns the "<prefix>/pkg/<module>" root for a package path
// under the pkg/<module> application-module layout, or ok=false if pkgPath has
// no /pkg/ segment (or nothing follows it).
func PkgModuleRoot(path string) (string, bool) {
	// The module.md application-module layout marker: pkg/<module>/ repeats the
	// same layered structure (dal/, domain/, server/) as internal/.
	const pkgSeg = "/pkg/"
	prefix, rest, ok := strings.Cut(path, pkgSeg)
	if !ok || rest == "" {
		return "", false
	}
	module, _, _ := strings.Cut(rest, "/")
	if module == "" {
		return "", false
	}
	return prefix + pkgSeg + module, true
}

// nonEmpty drops empty segments (from leading/trailing/duplicate slashes).
func nonEmpty(segs []string) []string {
	out := segs[:0]
	for _, s := range segs {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func matchAt(segs []string, i int, seq []string) bool {
	for j, s := range seq {
		if segs[i+j] != s {
			return false
		}
	}
	return true
}
