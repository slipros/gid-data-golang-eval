package optsstyle

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer — default scope: handler leaves and the domain service/usecase
// layers. The config layer (svc/config) and cross-package Options types
// (svc/extlib used from the handler) produce no diagnostics.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...")
}

// TestCustomScope — settings replace the default scope: here the rule is
// enforced only in "config" leaf packages.
func TestCustomScope(t *testing.T) {
	a := NewAnalyzer(Settings{
		Leaf: [][]string{{"config"}},
	})
	analysistest.Run(t, analysistest.TestData(), a, "custom/...")
}
