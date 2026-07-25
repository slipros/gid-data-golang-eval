package logctx_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/logctx"
)

// TestAnalyzer — the logrus stack: the context arrives through WithContext,
// the error through WithError.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), logctx.Analyzer, "logctx")
}

// TestAnalyzerSlog — the same rule on the slog stack, where the shapes differ:
// the context travels in the *Context variant of the terminal call and the
// error goes in as an attribute.
func TestAnalyzerSlog(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), logctx.Analyzer, "logctxslog")
}
