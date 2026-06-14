package entitymethod_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/entitymethod"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), entitymethod.Analyzer, "svc/...")
}

// TestExclude — methods from settings.exclude are not reported.
func TestExclude(t *testing.T) {
	a := entitymethod.NewAnalyzer(entitymethod.Settings{
		Exclude: []string{"Job.Close", "Ping"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "excluded/...")
}
