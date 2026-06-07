package filterplace_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/filterplace"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), filterplace.Analyzer,
		"dalsvc/...", "domainsvc/...", "httpsvc/...")
}
