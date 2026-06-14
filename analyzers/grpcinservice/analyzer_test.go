package grpcinservice_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/grpcinservice"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), grpcinservice.Analyzer, "svc/...")
}

// TestExclude — import paths from settings.exclude are allowed.
func TestExclude(t *testing.T) {
	a := grpcinservice.NewAnalyzer(grpcinservice.Settings{
		Exclude: []string{"excluded/pkg/api/orderpb"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "excluded/...")
}
