package constscope

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...", "plain/...")
}

// TestAnalyzerLibrary — in a library module exported constants are public API and
// are not reported, while an unexported single-use constant still is.
func TestAnalyzerLibrary(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "lib/...")
}

func TestAnalyzerExclude(t *testing.T) {
	a := NewAnalyzer(Settings{Exclude: []string{"LegacyExported"}})
	analysistest.Run(t, analysistest.TestData(), a, "excluded/...")
}
