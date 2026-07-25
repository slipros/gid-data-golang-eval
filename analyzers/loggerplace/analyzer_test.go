package loggerplace

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer — default suffixes (Options/Config/Settings) on testdata/src/svc:
//   - positive: a slog and a logrus logger inside options structs;
//   - negative: options without a logger, the logger arriving as a parameter;
//   - boundary: an interface-typed logger field, a struct without the suffix;
//   - non-applicability: an options struct of plain configuration.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc")
}

// TestAnalyzerSuffixes — settings.suffixes replaces the default list: the
// configured "Params" suffix is checked, "Options" is not.
func TestAnalyzerSuffixes(t *testing.T) {
	a := NewAnalyzer(Settings{Suffixes: []string{"Params"}})
	analysistest.Run(t, analysistest.TestData(), a, "custom")
}
