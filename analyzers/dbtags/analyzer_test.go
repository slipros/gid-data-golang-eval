package dbtags_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/dbtags"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), dbtags.Analyzer, "svc/...")
}

// TestCustomTags — settings.tags: для ClickHouse-библиотеки тег ch.
func TestCustomTags(t *testing.T) {
	a := dbtags.NewAnalyzer(dbtags.Settings{Tags: []string{"db", "ch"}})
	analysistest.Run(t, analysistest.TestData(), a, "clickhouse/...")
}

// TestModelAnalyzer — GID-168: запрет db-тегов у полей структур в /domain/**.
func TestModelAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), dbtags.ModelAnalyzer, "modeltags/...")
}
