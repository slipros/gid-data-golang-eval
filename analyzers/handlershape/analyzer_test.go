package handlershape_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/handlershape"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), handlershape.Analyzer, "svc/...")
}

// TestExclude — types from settings.exclude are not reported.
func TestExclude(t *testing.T) {
	a := handlershape.NewAnalyzer(handlershape.Settings{
		Exclude: []string{"HealthCheck", "Job"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "excluded/...")
}
