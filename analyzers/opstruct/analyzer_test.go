package opstruct_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/opstruct"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), opstruct.Analyzer,
		"domainsvc/...", "dalsvc/...", "clientsvc/...")
}
