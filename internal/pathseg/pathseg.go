// Package pathseg — Clean Architecture layer detection by the segments of a
// package's import path. The convention: a layer is defined by a sequence of
// segments, e.g. /dal/repository or /domain/model — regardless of the module
// prefix.
package pathseg

import (
	"strings"
	"sync"
)

// layerPairs — the canonical layer/sublayer pairs of a service (ARCHITECTURE.md,
// the same tree giddirtree keeps for internal/). A pair is the only marker of
// the layered structure strong enough to be read out of a bare import path: a
// lone segment named client or domain is an ordinary package name in any
// library, but dal/entity or domain/service is the service template speaking.
var layerPairs = map[string]map[string]struct{}{
	"dal":    {"entity": {}, "repository": {}},
	"domain": {"model": {}, "service": {}, "usecase": {}},
	"server": {"grpc": {}, "http": {}},
}

// layerSegmentsCache — import path -> its layer segments. Splitting the path
// allocates, and HasLayer is the hottest call in the rule set: every layer rule
// asks it for every package and, in the import rules, for every import of every
// file. The same handful of paths is asked about again and again within a run,
// so the split happens once per distinct path.
var layerSegmentsCache sync.Map

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
//
// The returned slice is shared between callers and must not be modified.
func LayerSegments(path string) []string {
	if v, ok := layerSegmentsCache.Load(path); ok {
		if segs, isSlice := v.([]string); isSlice {
			return segs
		}
	}

	segs := layerSegments(path)
	layerSegmentsCache.Store(path, segs)

	return segs
}

func layerSegments(path string) []string {
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

// PkgModuleRoot returns the module root for a package path under the pkg/
// application-module layout, or ok=false if path has no /pkg/ segment (or
// nothing follows it).
//
// A module is not always one segment deep: resource-registry groups its
// integrations by category and vendor (pkg/integration/push/firebase,
// pkg/integration/ad_cabinet/yandex_audience). Reading exactly one segment made
// every package of such a module layer-less — LayerSegments of
// .../firebase/domain/service came out as ["push","firebase","domain","service"],
// so HasLayer(…, "domain","service") was false — and that silently disabled
// every layer rule inside the whole module tree, which linted clean (incident
// 2026-08-06, resource-registry).
//
// So the root is taken to end where a canonical layer PAIR begins
// (.../firebase + domain/service). Without a pair the old one-segment root
// stands, which is what keeps a nested package that merely borrows a layer name
// from being read as that layer (pkg/billing/connect/client/x is not the client
// layer — the same false positive HasLayer fixed for internal/). The cost of
// that caution: a lone layer directory of a deeply nested module
// (pkg/integration/push/firebase/client, no sublayer under it) is still not
// recognised as a layer.
func PkgModuleRoot(path string) (string, bool) {
	// The module.md application-module layout marker: pkg/<module>/ repeats the
	// same layered structure (dal/, domain/, server/) as internal/.
	const pkgSeg = "/pkg/"
	prefix, rest, ok := strings.Cut(path, pkgSeg)
	if !ok || rest == "" {
		return "", false
	}
	segs := nonEmpty(Segments(rest))
	if len(segs) == 0 {
		return "", false
	}

	end := 1 // the pkg/<module> root, unless a layer pair says otherwise
	for i := 0; i+1 < len(segs); i++ {
		subs, isLayer := layerPairs[segs[i]]
		if _, isPair := subs[segs[i+1]]; isLayer && isPair {
			end = i
			break
		}
	}
	if end == 0 {
		// pkg/ itself holds the layers (pkg/domain/model) — pkg is the root.
		return strings.TrimSuffix(prefix+pkgSeg, "/"), true
	}

	return prefix + pkgSeg + strings.Join(segs[:min(end, len(segs))], "/"), true
}

// SameLibrary reports whether importPath is library (an import path without a
// version suffix) at any major version: "github.com/gofrs/uuid" matches both
// itself and "github.com/gofrs/uuid/v5". A rule that pins a library by its
// import path must go through this — a service on v5 is on the same library,
// and comparing the bare path silently makes such a rule a no-op.
// A subpackage ("github.com/gofrs/uuid/namespace") is not the library itself.
func SameLibrary(importPath, library string) bool {
	if importPath == library {
		return true
	}
	rest, ok := strings.CutPrefix(importPath, library+"/v")
	if !ok || rest == "" {
		return false
	}
	for _, r := range rest {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
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
