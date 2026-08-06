// Package moduleimports implements rule GID-259: an application module
// (pkg/<module>/**, module.md) owns its data layer — it must not reach into the
// core layers of the service it lives in.
//
// A module repeats the whole layered structure under its own root
// (pkg/integration/push/firebase/{client,dal,domain,server}), so its service has
// a dal of its own to speak to. What it shares with the core is the business
// vocabulary and the core's business layer: /domain/** (model, service,
// usecase) is allowed by default (settings.allow), everything else of the core
// — /dal/**, /client/**, /server/**, /event/** … — is not.
//
// The shape this closes (incident 2026-08-06, resource-registry
// pkg/integration/push/firebase/domain/service/integration.go): the module's
// service declared a CoreIntegrationRepository interface over the CORE entity
// (internal/dal/entity), so it read and wrote public.integration directly,
// through the core's data layer, in its own transaction. A module's service
// takes core data through a core SERVICE, and the two-writer orchestration
// belongs to a usecase (GID-260), not to a service.
//
// Judged are the business layers of the module — /domain/** and /dal/**
// (settings.layers). The module's dal is in scope for the same reason its
// service is: a repository of the module reaching for the core sentinels
// (commonentity.ErrNoResult, ErrAlreadyExists) works inside the core's data
// vocabulary instead of its own — the core's failures reach the module as
// DOMAIN errors of the core service/usecase, and a module usecase handles them
// from there (owner decision 2026-08-06).
//
// The transport layers of a module (/server/**, /client/**) are not judged by
// default: what they take from the core there is shared infrastructure — the
// i18n registrar of the validator, the http router wiring (lk-api) — not core
// data. A service that wants them judged adds them to settings.layers.
//
// The module root itself (pkg/<module>/module.go) is its composition root —
// wiring is exactly where the concrete core repository is allowed to be named,
// the same carve-out GID-241 makes for /app/**.
//
// A _test.go file is judged like production code: a test double of the CORE
// repository is the very dependency the rule forbids, so a test needing one is
// evidence of the violation, not scaffolding around it.
//
// Exceptions: //nolint:gidmoduleimports, or settings.exclude (an import path
// prefix) for a deliberate, listed sharing.
package moduleimports

import (
	"go/ast"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-259"

// defaultAllow — core layers a module may import.
var defaultAllow = []string{"domain"}

// defaultLayers — layers of the module whose packages are judged.
var defaultLayers = []string{"domain", "dal"}

// Analyzer — rule GID-259 with default settings.
var Analyzer = NewAnalyzer(Settings{})

// Settings — settings of rule GID-259 from .golangci.yml.
type Settings struct {
	// Allow — core layer folders a module may import (the first segment after
	// /internal/). Defaults to ["domain"] when empty.
	Allow []string `json:"allow"`
	// Layers — layers of the module whose packages are judged (the first
	// segment of the package's layer path). Defaults to ["domain", "dal"].
	Layers []string `json:"layers"`
	// Exclude — import paths (a full path or its prefix) shared deliberately.
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-259 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	allow := s.Allow
	if len(allow) == 0 {
		allow = defaultAllow
	}
	layers := s.Layers
	if len(layers) == 0 {
		layers = defaultLayers
	}

	return &analysis.Analyzer{
		Name: "gidmoduleimports",
		Doc: ruleID + ": an application module (pkg/<module>/**) must not import the core layers " +
			"besides /domain/**. Fix: use the module's own dal, and take core data through a core service",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, allow, layers, s.Exclude)
		},
	}
}

func run(pass *analysis.Pass, allow, layers, excl []string) (any, error) {
	pkgPath := pass.Pkg.Path()
	corePrefix, ok := moduleLayerPkg(pkgPath, layers)
	if !ok {
		return nil, nil
	}

	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, imp := range file.Imports {
			path, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				continue
			}
			layer, ok := coreLayer(path, corePrefix)
			if !ok || slices.Contains(allow, layer) || excluded(excl, path) {
				continue
			}
			pass.Reportf(imp.Pos(),
				"%s: module package %q must not import the core layer %q — a module owns its dal, "+
					"and only the core /domain/** is shared. Fix: declare the repository interface over "+
					"the module's own entity (<module>/dal/entity), and take core data through a core "+
					"service injected in module.go",
				ruleID, pkgPath, path)
		}
	}

	return nil, nil
}

// moduleLayerPkg reports whether pkgPath is a package inside a layer of an
// application module, and returns the prefix of the service the module lives in
// (the path before /pkg/) — the boundary that tells the core apart from a
// third-party library with an /internal/ segment.
//
// A package outside the judged layers is skipped: the module root holding
// module.go and a grouping directory above it (pkg/integration/push) hold no
// layer at all — wiring is where naming a concrete core repository is
// legitimate — and the transport layers are left to settings.layers.
func moduleLayerPkg(pkgPath string, layers []string) (string, bool) {
	const pkgSeg = "/pkg/"
	prefix, _, ok := strings.Cut(pkgPath, pkgSeg)
	if !ok {
		return "", false
	}
	segs := pathseg.LayerSegments(pkgPath)
	if len(segs) == 0 || !slices.Contains(layers, segs[0]) {
		return "", false
	}

	return prefix, true
}

// coreLayer returns the layer folder an import points at inside the core of the
// same service (<corePrefix>/internal/<layer>/...), or ok=false for anything
// else — a third-party library, another module, the module's own packages.
func coreLayer(importPath, corePrefix string) (string, bool) {
	const internalSeg = "/internal/"

	rest, ok := strings.CutPrefix(importPath, corePrefix+internalSeg)
	if !ok {
		return "", false
	}
	layer, _, _ := strings.Cut(rest, "/")
	if layer == "" {
		return "", false
	}

	return layer, true
}

// excluded reports whether the import path is listed in settings.exclude —
// as the path itself or as a prefix of it.
func excluded(excl []string, path string) bool {
	for _, e := range excl {
		if path == e || strings.HasPrefix(path, strings.TrimSuffix(e, "/")+"/") {
			return true
		}
	}

	return false
}
