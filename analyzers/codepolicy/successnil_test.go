package codepolicy

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer_SuccessNil(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), Analyzer, "successnil")
}
