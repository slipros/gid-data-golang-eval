package eventctor_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/eventctor"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), eventctor.Analyzer, "svc/...")
}

// TestExclude — constructors from settings.exclude are not reported.
func TestExclude(t *testing.T) {
	a := eventctor.NewAnalyzer(eventctor.Settings{
		Exclude: []string{"NewLegacyConsumer"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "excluded/...")
}

// TestLoggerTypes — settings.loggerTypes drives which parameter type counts as
// a logger: a custom allowlist accepts a project-specific type and rejects the
// defaults (slog.Logger here is not in the list).
func TestLoggerTypes(t *testing.T) {
	a := eventctor.NewAnalyzer(eventctor.Settings{
		LoggerTypes: []string{"mylog.Logger"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "custom/...")
}
