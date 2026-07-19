package convpolicy_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/convpolicy"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), convpolicy.Analyzer, "svc/...")
}

func TestAnalyzerExclude(t *testing.T) {
	a := convpolicy.NewAnalyzer(convpolicy.Settings{
		Exclude: []string{"asrFormatFromSource"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "exsvc/...")
}
