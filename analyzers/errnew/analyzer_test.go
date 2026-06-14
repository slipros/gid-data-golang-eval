package errnew_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/errnew"
)

// TestAnalyzer covers positive, negative and boundary cases.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), errnew.Analyzer, "svc/...")
}

// TestInapplicable — a package without github.com/pkg/errors is not reported.
func TestInapplicable(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), errnew.Analyzer, "nopkgerrors/...")
}
