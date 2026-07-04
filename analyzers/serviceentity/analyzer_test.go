package serviceentity_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/serviceentity"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), serviceentity.Analyzer, "svc/...")
}

// TestAnalyzerSettings — settings.suffixes adds a custom repository-name
// suffix on top of the default "Repository", settings.exclude skips a whole
// struct ("LegacySnapshot") or a single field ("Delivery.jobs").
func TestAnalyzerSettings(t *testing.T) {
	a := serviceentity.NewAnalyzer(serviceentity.Settings{
		Suffixes: []string{"Repository", "Store"},
		Exclude:  []string{"LegacySnapshot", "Delivery.jobs"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "custom/...")
}
