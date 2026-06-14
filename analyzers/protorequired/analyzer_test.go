package protorequired_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/protorequired"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), protorequired.Analyzer, "svc/...")
}

// TestExclude — fields from settings.exclude are not reported.
func TestExclude(t *testing.T) {
	a := protorequired.NewAnalyzer(protorequired.Settings{
		Exclude: []string{"CreateStageRequest.Executor"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "excluded/...")
}
