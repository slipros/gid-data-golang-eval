package failedto

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...")
}

// TestInapplicable — a package without github.com/pkg/errors is not reported.
func TestInapplicable(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "nopkgerrors/...")
}

// TestCustomPrefixes — settings.prefixes replaces the default list.
func TestCustomPrefixes(t *testing.T) {
	a := NewAnalyzer(Settings{Prefixes: []string{"oops"}})
	analysistest.Run(t, analysistest.TestData(), a, "custom/...")
}
