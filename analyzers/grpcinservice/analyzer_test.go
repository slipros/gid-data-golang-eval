package grpcinservice

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...")
}

// TestNoDataLayer — a module without a data layer (a BFF: /domain/service and
// /server/http and nothing below them) is not judged at all: the fix of the
// rule is "move the call into a repository", and there is no repository to
// move it into. The fixture imports grpc and a pb stub in /domain/service and
// carries no // want.
func TestNoDataLayer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "bff/...")
}

// TestExclude — import paths from settings.exclude are allowed.
func TestExclude(t *testing.T) {
	a := NewAnalyzer(Settings{
		Exclude: []string{"excluded/pkg/api/orderpb"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "excluded/...")
}
