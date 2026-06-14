package validatorshape_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/validatorshape"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), validatorshape.Analyzer, "svc/...")
}

// TestExclude — types from settings.exclude are not reported.
func TestExclude(t *testing.T) {
	a := validatorshape.NewAnalyzer(validatorshape.Settings{
		Exclude: []string{"HealthCheck"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "excluded/...")
}
