package validatorlib

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...")
}

// TestExclude — validate packages from settings.exclude are exempt.
func TestExclude(t *testing.T) {
	a := NewAnalyzer(Settings{
		Exclude: []string{"kafka/consumer/validate"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "excluded/...")
}
