package inout_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/inout"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), inout.Analyzer, "svc/...")
}

// TestExclude — methods from settings.exclude are not reported.
func TestExclude(t *testing.T) {
	a := inout.NewAnalyzer(inout.Settings{
		Exclude: []string{"Snapshot.SnapshotPtr"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "excluded/...")
}
