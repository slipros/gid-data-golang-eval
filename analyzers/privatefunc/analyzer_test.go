package privatefunc_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/privatefunc"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), privatefunc.Analyzer, "svc/...")
}
