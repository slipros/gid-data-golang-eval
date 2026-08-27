package modelplace

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "mp/...")
}

// TestAnalyzerSettings — settings.suffixes replaces the default options
// suffixes; settings.exclude (the "Struct" form) exempts a type in both parts.
func TestAnalyzerSettings(t *testing.T) {
	a := NewAnalyzer(Settings{
		Suffixes: []string{"Spec"},
		Exclude:  []string{"LegacyDTO", "Exempt"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "custom/...")
}
