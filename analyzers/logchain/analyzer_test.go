package logchain

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer — the logrus stack.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "logchain")
}

// TestAnalyzerSlog — the same rule on the slog stack (With instead of WithField).
func TestAnalyzerSlog(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "logchainslog")
}
