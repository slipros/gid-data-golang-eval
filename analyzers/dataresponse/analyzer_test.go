package dataresponse_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/dataresponse"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), dataresponse.Analyzer, "svc/...")
}

// TestExclude — handlers from settings.exclude are allowed.
func TestExclude(t *testing.T) {
	a := dataresponse.NewAnalyzer(dataresponse.Settings{
		Exclude: []string{"Health"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "excluded/...")
}
