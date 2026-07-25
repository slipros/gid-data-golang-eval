package constscope

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...", "plain/...")
}

func TestAnalyzerExclude(t *testing.T) {
	a := NewAnalyzer(Settings{Exclude: []string{"LegacyExported"}})
	analysistest.Run(t, analysistest.TestData(), a, "excluded/...")
}
