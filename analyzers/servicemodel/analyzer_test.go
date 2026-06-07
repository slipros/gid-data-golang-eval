package servicemodel_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/servicemodel"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), servicemodel.Analyzer, "svc/...")
}
