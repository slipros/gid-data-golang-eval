package closurector

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...")
}

// TestAnalyzerSettings — settings.exclude in both forms ("Type.Method" and
// "Function") and settings.option-suffixes replacing the default option types.
func TestAnalyzerSettings(t *testing.T) {
	a := NewAnalyzer(Settings{
		OptionSuffixes: []string{"Setting"},
		Exclude:        []string{"Sender.contactFilter", "statusFilter"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "custom/...")
}
