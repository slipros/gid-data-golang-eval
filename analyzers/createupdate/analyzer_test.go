package createupdate_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/createupdate"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), createupdate.Analyzer, "svc/...")
}

// TestExclude — методы из settings.exclude не репортятся.
func TestExclude(t *testing.T) {
	a := createupdate.NewAnalyzer(createupdate.Settings{
		Exclude: []string{"Job.CreateJob", "UpdateSession"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "excluded/...")
}
