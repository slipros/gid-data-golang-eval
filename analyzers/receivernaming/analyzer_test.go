package receivernaming_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/receivernaming"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), receivernaming.Analyzer, "svc/...")
}
