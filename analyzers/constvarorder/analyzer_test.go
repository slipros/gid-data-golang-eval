package constvarorder_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/constvarorder"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), constvarorder.Analyzer, "constvarorder")
}
