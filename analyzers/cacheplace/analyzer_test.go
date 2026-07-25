package cacheplace

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer — the default list of cache libraries.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...")
}

// TestCustomPackages — settings.packages replaces the default list.
func TestCustomPackages(t *testing.T) {
	a := NewAnalyzer(Settings{
		Packages: []string{"example.com/inhouse/cache"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "custom/...")
}
