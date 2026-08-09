package hastest

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"
)

// TestAnalyzer — the packages the rule must stay silent on, all of which
// analysistest can drive: they produce a base variant only.
//
//   - boundary: notests declares three candidates and no test file, and is NOT
//     reported — a pass without _test.go is indistinguishable from one whose
//     tests were withheld (run.tests: false);
//   - boundary: trivialonly exports getters, an enum String and a one-line
//     constructor only — nothing to report;
//   - non-applicability: the generated file, package main, /internal/app.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer,
		"notests", "trivialonly", "generated", "mainpkg", "svcmod/...")
}

// TestAnalyzerExcludePaths — non-applicability: a package under
// settings.exclude-paths is not checked at all.
func TestAnalyzerExcludePaths(t *testing.T) {
	a := NewAnalyzer(Settings{ExcludePaths: []string{"exempt/generated"}})
	analysistest.Run(t, analysistest.TestData(), a, "exempt/generated")
}

// TestTestVariant covers the packages that DO have tests. analysistest cannot:
// it hands the analyzer both variants of such a package, and the base one —
// which sees no _test.go and must therefore stay silent — would fail every
// `want` sitting in a production file. golangci-lint discards that variant
// (filterDuplicatePackages), so the run below is the one that matters.
//
//   - positive: svc.Segment.Rebuild is mentioned by no test;
//   - negative: Create is called by the test, Handed is handed over as a value,
//     MemRepo.Save is reached through the Repo interface, NewLimits runs through
//     the initializer of the DefaultLimits var the test names;
//   - boundary: the trivial getter, delegation and constructor of svc, and the
//     unexported normalize;
//   - non-applicability: NewFixture, an exported helper declared in
//     segment_test.go.
func TestTestVariant(t *testing.T) {
	got := runOnTestVariant(t, Analyzer, "svc")

	want := []string{
		"segment.go:54: GID-263: exported method Segment.Rebuild is not exercised by any test of this package. " +
			"Fix: add func TestSegment_Rebuild(t *testing.T) calling it",
	}
	if !slices.Equal(got, want) {
		t.Errorf("diagnostics:\ngot:  %q\nwant: %q", got, want)
	}
}

// TestTestVariantExclude — non-applicability: a method named on
// settings.exclude needs no test of its own. Without the setting the same
// package reports it, which is what the first half asserts.
func TestTestVariantExclude(t *testing.T) {
	got := runOnTestVariant(t, Analyzer, "exclcase")
	if len(got) != 1 || !strings.Contains(got[0], "Segment.Rebuild") {
		t.Fatalf("without the exclusion, want one Segment.Rebuild diagnostic, got %q", got)
	}

	a := NewAnalyzer(Settings{Exclude: []string{"Segment.Rebuild"}})
	if got := runOnTestVariant(t, a, "exclcase"); len(got) != 0 {
		t.Errorf("with the exclusion, want no diagnostics, got %q", got)
	}
}

// runOnTestVariant loads a testdata package the way analysistest does (GOPATH
// mode, Tests: true) and runs the analyzer over its "pkg [pkg.test]" variant —
// the one carrying the _test.go files. Diagnostics come back as
// "file:line: message", in source order.
func runOnTestVariant(t *testing.T, a *analysis.Analyzer, pattern string) []string {
	t.Helper()

	dir := analysistest.TestData()
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports |
			packages.NeedTypes | packages.NeedTypesSizes | packages.NeedSyntax | packages.NeedTypesInfo |
			packages.NeedDeps,
		Dir:   dir,
		Tests: true,
		Env:   append(os.Environ(), "GOPATH="+dir, "GO111MODULE=off", "GOWORK=off"),
	}

	pkgs, err := packages.Load(cfg, pattern)
	if err != nil {
		t.Fatalf("load %s: %v", pattern, err)
	}

	for _, pkg := range pkgs {
		// "svc [svc.test]" — the variant holding the package's own test files;
		// the synthesized "svc.test" binary and the base variant are not it.
		if !strings.HasSuffix(pkg.ID, ".test]") {
			continue
		}
		if len(pkg.Errors) > 0 {
			t.Fatalf("load %s: %v", pkg.ID, pkg.Errors)
		}

		var got []string
		pass := &analysis.Pass{
			Analyzer:  a,
			Fset:      pkg.Fset,
			Files:     pkg.Syntax,
			Pkg:       pkg.Types,
			TypesInfo: pkg.TypesInfo,
			ResultOf:  map[*analysis.Analyzer]any{inspect.Analyzer: inspector.New(pkg.Syntax)},
			Report: func(d analysis.Diagnostic) {
				pos := pkg.Fset.Position(d.Pos)
				got = append(got, fmt.Sprintf("%s:%d: %s", filepath.Base(pos.Filename), pos.Line, d.Message))
			},
		}

		if _, err := a.Run(pass); err != nil {
			t.Fatalf("run %s: %v", pkg.ID, err)
		}

		return got
	}

	t.Fatalf("no test variant of %q was loaded", pattern)

	return nil
}
