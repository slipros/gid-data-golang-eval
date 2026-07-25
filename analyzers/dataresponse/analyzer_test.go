package dataresponse

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...")
}

// TestExclude — handlers from settings.exclude are allowed.
func TestExclude(t *testing.T) {
	a := NewAnalyzer(Settings{
		Exclude: []string{"Health"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "excluded/...")
}
