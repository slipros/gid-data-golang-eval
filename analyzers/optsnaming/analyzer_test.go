package optsnaming_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/optsnaming"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), optsnaming.Analyzer, "svc/...")
}
