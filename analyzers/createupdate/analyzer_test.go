package createupdate

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...")
}

// TestExclude — methods from settings.exclude are not reported.
func TestExclude(t *testing.T) {
	a := NewAnalyzer(Settings{
		Exclude: []string{"Job.CreateJob", "UpdateSession"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "excluded/...")
}
