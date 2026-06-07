package chainperline_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/chainperline"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), chainperline.Analyzer, "chains/...")
}

func TestAnalyzerMinCalls(t *testing.T) {
	a := chainperline.NewAnalyzer(chainperline.Settings{MinCalls: 3})
	analysistest.Run(t, analysistest.TestData(), a, "threshold/...")
}
