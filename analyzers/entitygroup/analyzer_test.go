package entitygroup_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/entitygroup"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), entitygroup.Analyzer, "entitygroup")
}
