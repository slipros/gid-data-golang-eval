package errname

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...")
}

func TestAnalyzerCustomSettings(t *testing.T) {
	a := NewAnalyzer(Settings{
		Names:   []string{"ErrOops", "ErrLegacy"},
		Exclude: []string{"ErrLegacy"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "custom/...")
}
