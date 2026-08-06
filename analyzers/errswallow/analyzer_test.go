package errswallow

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer,
		"svc/internal/domain/service", "svc/internal/server/grpc")
}

// TestExclude — a function on settings.exclude may swallow the error: soft
// degradation as a stated functional requirement. The fixture holds the same
// violation as the domain service above and stays clean only because of the
// setting.
func TestExclude(t *testing.T) {
	a := NewAnalyzer(Settings{Exclude: []string{"Resolver.degradeSilently"}})
	analysistest.Run(t, analysistest.TestData(), a, "excluded/internal/domain/service")
}
