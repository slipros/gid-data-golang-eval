package errnew

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer covers positive, negative and boundary cases.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...")
}

// TestInapplicable — a package without github.com/pkg/errors is not reported.
func TestInapplicable(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "nopkgerrors/...")
}
