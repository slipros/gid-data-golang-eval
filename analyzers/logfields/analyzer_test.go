package logfields

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer — the logrus stack: repeated WithField in one chain.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "logfields")
}

// TestAnalyzerSlog — the slog stack is out of scope: With takes variadic pairs.
func TestAnalyzerSlog(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "logfieldsslog")
}
