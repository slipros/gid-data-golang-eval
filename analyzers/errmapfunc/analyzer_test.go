package errmapfunc_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/errmapfunc"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), errmapfunc.Analyzer, "svc")
}
