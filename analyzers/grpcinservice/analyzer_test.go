package grpcinservice

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...")
}

// TestExclude — import paths from settings.exclude are allowed.
func TestExclude(t *testing.T) {
	a := NewAnalyzer(Settings{
		Exclude: []string{"excluded/pkg/api/orderpb"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "excluded/...")
}
