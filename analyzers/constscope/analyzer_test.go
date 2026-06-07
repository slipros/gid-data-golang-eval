package constscope_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/constscope"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), constscope.Analyzer, "svc/...", "plain/...")
}

func TestAnalyzerExclude(t *testing.T) {
	a := constscope.NewAnalyzer(constscope.Settings{Exclude: []string{"LegacyExported"}})
	analysistest.Run(t, analysistest.TestData(), a, "excluded/...")
}
