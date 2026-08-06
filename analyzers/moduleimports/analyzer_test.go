package moduleimports

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...")
}

// TestAnalyzerSettings — settings.allow widens the shared core layers (a
// project sharing its core client on purpose), settings.layers adds the
// module's transport to the judged layers, settings.exclude clears a single
// import path.
func TestAnalyzerSettings(t *testing.T) {
	a := NewAnalyzer(Settings{
		Allow:   []string{"domain", "client"},
		Layers:  []string{"domain", "dal", "server"},
		Exclude: []string{"custom/internal/dal/repository"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "custom/...")
}
