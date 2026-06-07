package ifacemin_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/ifacemin"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), ifacemin.Analyzer, "svc/...")
}

func TestAnalyzerExclude(t *testing.T) {
	a := ifacemin.NewAnalyzer(ifacemin.Settings{
		Exclude: []string{"LegacyGateway", "AlertSink.Flush"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "exsvc/...")
}
