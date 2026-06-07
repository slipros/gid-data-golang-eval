package initclean_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/initclean"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), initclean.Analyzer, "bad", "good", "noinit")
}
