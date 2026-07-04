// Package convpure implements rule GID-235 (slug convert-purity, linter
// gidconvpure). Source: converter.md style guide — "Converters are functions
// without side effects"; a converter operates only on vocabulary types.
//
// Scope: packages whose import path ends with the "convert" segment
// (pathseg.EndsWith(pkgPath, "convert")) — the converter package itself, not
// a package that merely contains "convert" as a substring of its last
// segment (e.g. xconvert) and not a package where "convert" sits in the
// middle of the path (e.g. convert/util).
//
// For same-module imports (module boundary as in layerimports: the
// /internal/ segment for the canonical layout, otherwise the first path
// segment) the following segments are forbidden: domain/service,
// domain/usecase, dal/repository, metric, app, server, schedule, validate,
// event. The event/dto exception is checked first: event/dto is itself a
// vocabulary package (the event DTO) and is allowed even though it is
// nested under the otherwise-banned event segment.
//
// Allowed (not checked): stdlib, third-party modules (except the banned list
// below), and same-module vocabulary packages — domain/model/**,
// dal/entity/**, client/**, event/dto, genproto/pb.
//
// Third-party imports are checked against settings.packages (exact path or
// path prefix), which replaces the default list: ["github.com/sirupsen/logrus"]
// — a converter must not log, since it is a pure function.
//
// Escape hatch: //nolint:gidconvpure.
package convpure

import (
	"go/ast"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-235"

// bannedLayers — same-module segments forbidden to a convert package:
// business layers and other packages that carry side effects. Each entry is
// checked independently; the event/dto exception (see checkSameModule) is
// evaluated before the "event" ban, so event/dto itself is never flagged.
var bannedLayers = [][]string{
	{"domain", "service"},
	{"domain", "usecase"},
	{"dal", "repository"},
	{"metric"},
	{"app"},
	{"server"},
	{"schedule"},
	{"validate"},
	{"event"},
}

// defaultThirdParty — side-effect-bearing third-party libraries forbidden to
// a convert package (import path prefixes).
var defaultThirdParty = []string{
	"github.com/sirupsen/logrus",
}

// Analyzer — GID-235 with the default third-party ban list.
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Packages — side-effect-bearing third-party libraries (import path
	// prefixes) forbidden to a convert package. Replaces the default list.
	Packages []string `json:"packages"`
}

// NewAnalyzer builds the GID-235 analyzer from the linter settings.
func NewAnalyzer(s Settings) *analysis.Analyzer {
	thirdParty := s.Packages
	if len(thirdParty) == 0 {
		thirdParty = defaultThirdParty
	}
	return &analysis.Analyzer{
		Name: "gidconvpure",
		Doc: ruleID + ": a convert package must not import business-layer or " +
			"side-effect packages — a converter is a pure function over " +
			"vocabulary types (model/entity/dto/client/pb)",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, thirdParty)
		},
	}
}

func run(pass *analysis.Pass, thirdParty []string) (any, error) {
	if !pathseg.EndsWith(pass.Pkg.Path(), "convert") {
		return nil, nil
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		checkImports(pass, file, thirdParty)
	}
	return nil, nil
}

func checkImports(pass *analysis.Pass, file *ast.File, thirdParty []string) {
	pkgPath := pass.Pkg.Path()
	for _, imp := range file.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}
		if sameModule(pkgPath, path) {
			checkSameModule(pass, imp, path)
			continue
		}
		checkThirdParty(pass, imp, path, thirdParty)
	}
}

// checkSameModule reports the first banned layer matched by path, unless
// path falls under the event/dto exception.
func checkSameModule(pass *analysis.Pass, imp *ast.ImportSpec, path string) {
	if pathseg.Contains(path, "event", "dto") {
		return
	}
	for _, banned := range bannedLayers {
		if !pathseg.Contains(path, banned...) {
			continue
		}
		report(pass, imp, path)
		return
	}
}

func checkThirdParty(pass *analysis.Pass, imp *ast.ImportSpec, path string, thirdParty []string) {
	for _, pkg := range thirdParty {
		if path == pkg || strings.HasPrefix(path, pkg+"/") {
			report(pass, imp, path)
			return
		}
	}
}

func report(pass *analysis.Pass, imp *ast.ImportSpec, path string) {
	pass.Reportf(imp.Pos(),
		"%s: convert package %q must not import %q — a converter is a pure "+
			"function over vocabulary types (model/entity/dto/client/pb); "+
			"business logic and side effects live in their layers",
		ruleID, pass.Pkg.Path(), path)
}

// sameModule tells whether an import belongs to the same module as the
// importing package (convention as in layerimports): for the canonical
// layout the /internal/ segment is the module boundary, otherwise (testdata,
// non-standard layout) the first path segment is compared.
func sameModule(pkgPath, importPath string) bool {
	const internalSeg = "/internal/"
	if module, _, ok := strings.Cut(pkgPath, internalSeg); ok {
		return strings.HasPrefix(importPath, module+internalSeg)
	}
	return firstSegment(pkgPath) == firstSegment(importPath)
}

func firstSegment(path string) string {
	seg, _, _ := strings.Cut(path, "/")
	return seg
}
