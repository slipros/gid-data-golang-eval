package modelmethod

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...", "dalsvc/...")
}

func TestAnalyzerExclude(t *testing.T) {
	a := NewAnalyzer(Settings{
		Exclude: []string{"legacyTitle", "Service.legacyRender"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "exsvc/...")
}
