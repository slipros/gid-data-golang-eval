package serviceorch

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...")
}

// TestAnalyzerSettings — settings.suffixes adds a custom repository-name suffix
// on top of the default "Repository"; settings.exclude clears a whole struct
// ("LegacyWriter") or a single field ("Mixed.tx").
func TestAnalyzerSettings(t *testing.T) {
	a := NewAnalyzer(Settings{
		Suffixes: []string{"Repository", "Store"},
		Exclude:  []string{"LegacyWriter", "Mixed.tx"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "custom/...")
}
