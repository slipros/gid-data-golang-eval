package eventctor_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/eventctor"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), eventctor.Analyzer, "svc/...")
}

// TestExclude — constructors from settings.exclude are not reported.
func TestExclude(t *testing.T) {
	a := eventctor.NewAnalyzer(eventctor.Settings{
		Exclude: []string{"NewLegacyConsumer"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "excluded/...")
}
