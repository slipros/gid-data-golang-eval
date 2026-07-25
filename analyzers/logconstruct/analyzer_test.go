package logconstruct

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer — the logrus stack: the entity is named with WithField.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "logconstruct")
}

// TestAnalyzerLayerKey — the naming pair: the key is the layer the entity
// lives in ("client" in /client/**, "service" in /domain/service), the value
// the entity name in lower snake_case.
func TestAnalyzerLayerKey(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer,
		"svclayer/internal/client/debugsink",
		"svclayer/internal/domain/service",
	)
}

// TestAnalyzerCompositionRoot — non-applicability: package main and the app
// layer are exempt; wiring hands the logger down to components that name
// themselves.
func TestAnalyzerCompositionRoot(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svcapp/internal/app/api")
}

// TestAnalyzerSlog — the same rule on the slog stack, where the entity is
// named with With("<entity>", <name>) instead.
func TestAnalyzerSlog(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "logconstructslog")
}
