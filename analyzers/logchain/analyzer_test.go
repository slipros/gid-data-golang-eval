package logchain_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/logchain"
)

// TestAnalyzer — the logrus stack.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), logchain.Analyzer, "logchain")
}

// TestAnalyzerSlog — the same rule on the slog stack (With instead of WithField).
func TestAnalyzerSlog(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), logchain.Analyzer, "logchainslog")
}
