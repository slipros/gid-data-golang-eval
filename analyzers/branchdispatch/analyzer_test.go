package branchdispatch

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...")
}

// TestAnalyzerSettings — settings.exclude clears a method by Type.Method
// ("Integration.Get") or by a bare method name ("Legacy").
func TestAnalyzerSettings(t *testing.T) {
	a := NewAnalyzer(Settings{Exclude: []string{"Integration.Get", "Legacy"}})
	analysistest.Run(t, analysistest.TestData(), a, "custom/...")
}
