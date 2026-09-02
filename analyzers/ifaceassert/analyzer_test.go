package ifaceassert

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer — the four case classes: positive (the package wires the very
// type it asserts — through a call argument, a composite literal field, an
// assignment and a declaration), negative (nothing in the package converts the
// type: a library, a container), boundary (a value asserted against a pointer
// conversion and back, a wider interface, an assertion of the empty interface,
// a conversion that lives only in _test.go), non-applicability (generated code,
// an interface asserted against an interface).
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...")
}

// TestExclude — an interface on settings.exclude keeps its assertion, by the
// interface name alone or qualified by the asserted type; the rest of the
// package is still judged.
func TestExclude(t *testing.T) {
	a := NewAnalyzer(Settings{Exclude: []string{"Store", "rowReader.Parquet"}})
	analysistest.Run(t, analysistest.TestData(), a, "excluded/...")
}
