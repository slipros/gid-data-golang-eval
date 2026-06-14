package validatorlib_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/validatorlib"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), validatorlib.Analyzer, "svc/...")
}

// TestExclude — validate packages from settings.exclude are exempt.
func TestExclude(t *testing.T) {
	a := validatorlib.NewAnalyzer(validatorlib.Settings{
		Exclude: []string{"kafka/consumer/validate"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "excluded/...")
}
