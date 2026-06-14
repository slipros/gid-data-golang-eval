package failedto_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/failedto"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), failedto.Analyzer, "svc/...")
}

// TestInapplicable — a package without github.com/pkg/errors is not reported.
func TestInapplicable(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), failedto.Analyzer, "nopkgerrors/...")
}

// TestCustomPrefixes — settings.prefixes replaces the default list.
func TestCustomPrefixes(t *testing.T) {
	a := failedto.NewAnalyzer(failedto.Settings{Prefixes: []string{"oops"}})
	analysistest.Run(t, analysistest.TestData(), a, "custom/...")
}
