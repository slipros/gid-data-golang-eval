package eventctor

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...")
}

// TestExclude — constructors from settings.exclude are not reported.
func TestExclude(t *testing.T) {
	a := NewAnalyzer(Settings{
		Exclude: []string{"NewLegacyConsumer"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "excluded/...")
}

// TestLoggerTypes — settings.loggerTypes drives which parameter type counts as
// a logger: a custom allowlist accepts a project-specific type and rejects the
// defaults (slog.Logger here is not in the list).
func TestLoggerTypes(t *testing.T) {
	a := NewAnalyzer(Settings{
		LoggerTypes: []string{"mylog.Logger"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "custom/...")
}
