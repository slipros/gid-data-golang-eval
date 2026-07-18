package approot_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/approot"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), approot.Analyzer, "approotsvc/...")
}

func TestAnalyzerExclude(t *testing.T) {
	a := approot.NewAnalyzer(approot.Settings{
		Exclude:      []string{"LegacyAdapter"},
		ExcludePaths: []string{"legacy"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "excludesvc/...")
}
