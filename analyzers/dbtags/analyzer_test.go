package dbtags_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/dbtags"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), dbtags.Analyzer, "svc/...")
}

// TestCustomTags — settings.tags: the ch tag for the ClickHouse library.
func TestCustomTags(t *testing.T) {
	a := dbtags.NewAnalyzer(dbtags.Settings{Tags: []string{"db", "ch"}})
	analysistest.Run(t, analysistest.TestData(), a, "clickhouse/...")
}

// TestModelAnalyzer — GID-168: a ban on db tags on struct fields in /domain/**.
func TestModelAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), dbtags.ModelAnalyzer, "modeltags/...", "nesteddomain/...")
}
