package logconstruct_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/logconstruct"
)

// TestAnalyzer — the logrus stack: the entity is named with WithField.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), logconstruct.Analyzer, "logconstruct")
}

// TestAnalyzerSlog — the same rule on the slog stack, where the entity is
// named with With("<entity>", <name>) instead.
func TestAnalyzerSlog(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), logconstruct.Analyzer, "logconstructslog")
}
