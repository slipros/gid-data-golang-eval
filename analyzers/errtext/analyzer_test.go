package errtext

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc")
}

// TestExclude — a function on settings.exclude may flatten the error (a
// functional requirement, legacy that cannot be rewritten now). The fixture
// holds the same violation as svc/client.go and stays clean only because of
// the setting.
func TestExclude(t *testing.T) {
	a := NewAnalyzer(Settings{Exclude: []string{"Client.legacyConfirm"}})
	analysistest.Run(t, analysistest.TestData(), a, "excluded")
}
