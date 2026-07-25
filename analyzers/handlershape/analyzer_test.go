package handlershape

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...")
}

// TestExclude — types from settings.exclude are not reported.
func TestExclude(t *testing.T) {
	a := NewAnalyzer(Settings{
		Exclude: []string{"HealthCheck", "Job"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "excluded/...")
}
