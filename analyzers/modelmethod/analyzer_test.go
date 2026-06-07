package modelmethod_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/modelmethod"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), modelmethod.Analyzer, "svc/...", "dalsvc/...")
}

func TestAnalyzerExclude(t *testing.T) {
	a := modelmethod.NewAnalyzer(modelmethod.Settings{
		Exclude: []string{"legacyTitle", "Service.legacyRender"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "exsvc/...")
}
