package mapout

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc")
}

// TestExclude — a function on settings.exclude may fill a map parameter
// (legacy that cannot be rewritten right now). The fixture holds the same
// violation as svc/resolver.go and stays clean only because of the setting.
func TestExclude(t *testing.T) {
	a := NewAnalyzer(Settings{Exclude: []string{"Resolver.legacyResolveChunk"}})
	analysistest.Run(t, analysistest.TestData(), a, "excluded")
}
