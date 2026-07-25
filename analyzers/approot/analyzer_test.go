package approot

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "approotsvc/...")
}

func TestAnalyzerExclude(t *testing.T) {
	a := NewAnalyzer(Settings{
		Exclude:      []string{"LegacyAdapter"},
		ExcludePaths: []string{"legacy"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "excludesvc/...")
}
