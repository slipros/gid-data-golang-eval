package enumplace_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/enumplace"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), enumplace.Analyzer,
		"dalsvc/...", "domainsvc/...", "nesteddal/...")
}
