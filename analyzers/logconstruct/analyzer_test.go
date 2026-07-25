package logconstruct

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer — the logrus stack: the entity is named with WithField.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "logconstruct")
}

// TestAnalyzerSlog — the same rule on the slog stack, where the entity is
// named with With("<entity>", <name>) instead.
func TestAnalyzerSlog(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "logconstructslog")
}
