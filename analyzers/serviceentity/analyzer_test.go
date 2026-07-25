package serviceentity

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...")
}

// TestAnalyzerSettings — settings.suffixes adds a custom repository-name
// suffix on top of the default "Repository", settings.exclude skips a whole
// struct ("LegacySnapshot") or a single field ("Delivery.jobs").
func TestAnalyzerSettings(t *testing.T) {
	a := NewAnalyzer(Settings{
		Suffixes: []string{"Repository", "Store"},
		Exclude:  []string{"LegacySnapshot", "Delivery.jobs"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "custom/...")
}
