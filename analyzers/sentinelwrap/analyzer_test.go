package sentinelwrap_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/sentinelwrap"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), sentinelwrap.Analyzer, "svc/...")
}

func TestAnalyzerExclude(t *testing.T) {
	a := sentinelwrap.NewAnalyzer(sentinelwrap.Settings{
		Exclude: []string{"Repo.excludedMethod"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "excludesvc/...")
}
