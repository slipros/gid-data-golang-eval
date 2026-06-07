package eventctor_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/eventctor"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), eventctor.Analyzer, "svc/...")
}

// TestExclude — конструкторы из settings.exclude не репортятся.
func TestExclude(t *testing.T) {
	a := eventctor.NewAnalyzer(eventctor.Settings{
		Exclude: []string{"NewLegacyConsumer"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "excluded/...")
}
