package scanrow_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/scanrow"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), scanrow.Analyzer, "svc/...")
}

func TestAnalyzerExclude(t *testing.T) {
	a := scanrow.NewAnalyzer(scanrow.Settings{
		Exclude: []string{"Repo.excludedMethod"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "excludesvc/...")
}
