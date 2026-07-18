// Package approot implements GID-246 (gidaproot): a struct type whose name
// carries "adapter" is a code smell — usually needless indirection.
//
// Agents love to spawn adapters: during wiring they wrap a dependency in a
// FooAdapter struct that adds nothing (the incident that motivated this rule —
// govorun-server's app/api/wiring filled up with such adapters). But an adapter
// only converts data between layers, and conversion is not a standalone artifact:
// it is the job of the layer's convert subpackage (e.g. domain/service/convert or
// dal/repository/convert), which maps model ↔ entity / model ↔ client models.
//
//	// bad — a standalone adapter struct that just maps types
//	type DedupAdapter struct{ cache *dedup.Cache }
//	func (d *DedupAdapter) GetResult(ctx context.Context, k string) (model.Result, error) {
//		res, err := d.cache.Lookup(ctx, k)
//		return model.Result{ID: res.ID, Hits: res.Hits}, err
//	}
//
//	// good — the mapping lives in the layer's convert subpackage
//	// dal/repository/convert:
//	func ToModelResult(res dedup.Result) model.Result { ... }
//
// Flagged: a struct type declaration whose name contains "adapter"
// (case-insensitive substring, so DedupAdapter, adapterImpl, HTTPAdapterV2 all
// match). Interfaces, type aliases and func types are not flagged — an interface
// named Adapter is a consumer-side port, which is legitimate. Generated code and
// _test.go files (mocks, stubs) are skipped.
//
// Legitimate adapters (e.g. an infrastructure adapter in internal/client) are
// exempted by directory via settings.exclude-paths, or by type name via
// settings.exclude.
package approot

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-246"

// Analyzer — GID-246 with default settings.
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Exclude — type names exempted from the rule.
	Exclude []string `json:"exclude"`
	// ExcludePaths — "/"-joined path-segment sequences; a package whose import
	// path contains such a sequence is skipped (e.g. "internal/client" keeps
	// legitimate infrastructure adapters alive).
	ExcludePaths []string `json:"exclude-paths"`
}

// NewAnalyzer builds the GID-246 analyzer.
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidapproot",
		Doc: ruleID + ": a struct named *Adapter is usually needless indirection — " +
			"adapt inline where the dependency is used, not via a standalone Adapter struct",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, s)
		},
	}
}

func run(pass *analysis.Pass, s Settings) (any, error) {
	if excludedPath(pass.Pkg.Path(), s.ExcludePaths) {
		return nil, nil
	}

	for _, file := range sourceFiles(pass) {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || !isStructType(ts) {
					continue
				}
				name := ts.Name.Name
				if !nameHasMarker(name) {
					continue
				}
				if exclude.Match(s.Exclude, name, name) {
					continue
				}
				pass.Report(analysis.Diagnostic{
					Pos: ts.Pos(),
					Message: fmt.Sprintf(
						"%s: %q is an adapter struct — an adapter only converts data between layers, "+
							"and conversion belongs in the layer's convert subpackage. "+
							"Fix: drop the adapter and move the mapping into <layer>/convert "+
							"(e.g. domain/service/convert or dal/repository/convert)",
						ruleID, name),
				})
			}
		}
	}
	return nil, nil
}

// sourceFiles returns the package's source files, skipping generated and _test.go.
func sourceFiles(pass *analysis.Pass) []*ast.File {
	var out []*ast.File
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		name := pass.Fset.Position(file.Pos()).Filename
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		out = append(out, file)
	}
	return out
}

// nameHasMarker reports whether name contains "adapter" (case-insensitive).
func nameHasMarker(name string) bool {
	return strings.Contains(strings.ToLower(name), "adapter")
}

// excludedPath reports whether pkgPath contains any of the excluded segment
// sequences (each entry is a "/"-joined sequence, e.g. "internal/client").
func excludedPath(pkgPath string, excludes []string) bool {
	for _, e := range excludes {
		seq := pathseg.Segments(strings.Trim(e, "/"))
		if pathseg.Contains(pkgPath, seq...) {
			return true
		}
	}
	return false
}

// isStructType reports whether ts declares a struct type (not an alias, enum,
// interface, or func type).
func isStructType(ts *ast.TypeSpec) bool {
	if ts.Assign.IsValid() {
		return false
	}
	_, ok := ts.Type.(*ast.StructType)
	return ok
}
