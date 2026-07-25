package convpolicy

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...")
}

func TestAnalyzerExclude(t *testing.T) {
	a := NewAnalyzer(Settings{
		Exclude: []string{"asrFormatFromSource"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "exsvc/...")
}
