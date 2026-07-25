package chainperline

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "chains/...")
}

func TestAnalyzerMinCalls(t *testing.T) {
	a := NewAnalyzer(Settings{MinCalls: 3})
	analysistest.Run(t, analysistest.TestData(), a, "threshold/...")
}
