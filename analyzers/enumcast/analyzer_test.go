package enumcast_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/enumcast"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), enumcast.Analyzer, "svc/...")
}
