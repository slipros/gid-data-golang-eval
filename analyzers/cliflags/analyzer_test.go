package cliflags_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/cliflags"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), cliflags.Analyzer, "svc/...")
}

// TestExclude — flag names from settings.exclude are exempt from GID-239.
func TestExclude(t *testing.T) {
	a := cliflags.NewAnalyzer(cliflags.Settings{
		Exclude: []string{"legacy-mode"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "custom/...")
}
