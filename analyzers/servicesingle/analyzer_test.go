package servicesingle_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/servicesingle"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), servicesingle.Analyzer, "svc/...")
}
