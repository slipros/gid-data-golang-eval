package enumstring

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "enumstring")
}

func TestBasedAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), BasedAnalyzer, "enumbased/...")
}
