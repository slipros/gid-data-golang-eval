package dbtags

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...")
}

// TestCustomTags — settings.tags: the ch tag for the ClickHouse library.
func TestCustomTags(t *testing.T) {
	a := NewAnalyzer(Settings{Tags: []string{"db", "ch"}})
	analysistest.Run(t, analysistest.TestData(), a, "clickhouse/...")
}

// TestModelAnalyzer — GID-168: a ban on db tags on struct fields in /domain/**.
func TestModelAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), ModelAnalyzer, "modeltags/...", "nesteddomain/...")
}
