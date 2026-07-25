// Package testpackage implements rule GID-250 (slug test-same-package, linter
// gidtestpackage): a test lives in the same package as the code it tests.
//
// A file "foo_test.go" declaring "package foo_test" is an external test
// package: it sees only the exported API, so a unit test of an unexported
// function has to be written through the public surface or moved. The team
// convention is the opposite of the golangci-lint "testpackage" linter (which
// is disabled in the config for exactly this reason): the test declares
// "package foo" and reaches whatever it needs directly.
//
// Detection is by the package clause of a _test.go file — the package name
// ending in "_test". Non-test files are never flagged (a package literally
// named "*_test" cannot exist outside test files).
//
// Escape hatch: //nolint:gidtestpackage on the package clause, or
// settings.exclude-paths for a directory that must stay external (a black-box
// suite over the public API).
package testpackage

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-250"

// Analyzer — GID-250 with default settings (no exclusions).
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// ExcludePaths — "/"-joined path-segment sequences; a package whose
	// import path contains such a sequence is not checked (e.g.
	// "test/blackbox"). An escape hatch for a deliberate black-box suite.
	ExcludePaths []string `json:"exclude-paths"`
}

// NewAnalyzer builds the GID-250 analyzer.
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidtestpackage",
		Doc: ruleID + ": a test lives in the same package as the code under test. " +
			"Fix: declare package <pkg> in the _test.go file instead of package <pkg>_test",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, s.ExcludePaths)
		},
	}
}

func run(pass *analysis.Pass, excludePaths []string) (any, error) {
	name := pass.Pkg.Name()
	if !strings.HasSuffix(name, "_test") {
		return nil, nil
	}
	// The package path of an external test package carries the same "_test"
	// suffix (".../blackbox_test"); exclude-paths is written against the real
	// directory, so the suffix is trimmed before matching.
	if excludedPath(strings.TrimSuffix(pass.Pkg.Path(), "_test"), excludePaths) {
		return nil, nil
	}
	inner := strings.TrimSuffix(name, "_test")
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		pass.Reportf(file.Name.Pos(),
			"%s: an external test package %q keeps the test away from the unexported code it tests. "+
				"Fix: declare %q in this file — a test lives in the same package as the code under test",
			ruleID, name, inner)
	}
	return nil, nil
}

// excludedPath reports whether the package is exempted through
// settings.exclude-paths (a "/"-joined sequence of path segments).
func excludedPath(pkgPath string, excludes []string) bool {
	for _, e := range excludes {
		if pathseg.Contains(pkgPath, pathseg.Segments(strings.Trim(e, "/"))...) {
			return true
		}
	}
	return false
}
