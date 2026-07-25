package cliflags

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...")
}

// TestExclude — flag names from settings.exclude are exempt from GID-239.
func TestExclude(t *testing.T) {
	a := NewAnalyzer(Settings{
		Exclude: []string{"legacy-mode"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "custom/...")
}
