package errmapfunc

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc")
}

// TestCustomPackages — a project-configured errors facade (settings.packages)
// is recognized as a classifier package: the facade mapper is flagged, the
// facade bool-predicate stays clean. Under the default whitelist the same
// package produces no diagnostics (myerrors is neither "errors" nor
// github.com/pkg/errors) — proving the setting, not a hardcoded list, drives it.
func TestCustomPackages(t *testing.T) {
	a := NewAnalyzer(Settings{Packages: []string{"myerrors"}})
	analysistest.Run(t, analysistest.TestData(), a, "customfacade")
}

// TestExclude — settings.exclude clears a framework-mandated converter
// (ValidationErrorConverter, the gdgrpcserver.WithErrorConverters shape — see
// the package doc) that would otherwise be flagged as a mapper. Converter, in
// the same fixture with the identical shape but not on the list, proves the
// setting drives the exclusion rather than a blanket exemption for the shape.
func TestExclude(t *testing.T) {
	a := NewAnalyzer(Settings{Exclude: []string{"ValidationErrorConverter"}})
	analysistest.Run(t, analysistest.TestData(), a, "excluded")
}
