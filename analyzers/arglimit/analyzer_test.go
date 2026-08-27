package arglimit

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer — the four case classes: positive (4+ arguments reported),
// negative (0..3 arguments pass), boundary (exactly 3 passes, exactly 4 is
// reported; 3 + ctx stays 3), non-applicability (a constructor, a _test.go
// file, a package outside /domain/**). Both layer layouts are exercised:
// internal/domain/... and pkg/<module>/domain/...
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...")
}

// TestExclude — a function or method on settings.exclude keeps its many
// arguments; the rest of the package is still judged.
func TestExclude(t *testing.T) {
	a := NewAnalyzer(Settings{Exclude: []string{"legacyConvert", "Converter.legacyMethod"}})
	analysistest.Run(t, analysistest.TestData(), a, "excluded/...")
}

// TestMaxArgs — the threshold is configurable: max-args 1 allows a single
// argument, two is already a violation.
func TestMaxArgs(t *testing.T) {
	a := NewAnalyzer(Settings{MaxArgs: 1})
	analysistest.Run(t, analysistest.TestData(), a, "maxargs/...")
}
