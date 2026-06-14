package fsmmap_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/fsmmap"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), fsmmap.Analyzer,
		"domainsvc/...", "othersvc/...")
}
