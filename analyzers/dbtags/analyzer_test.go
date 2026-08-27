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

// TestNoSQLStack — the applicability gate: a module whose /dal/repository
// speaks gRPC and fills entities from protobuf has no database to map onto, so
// GID-125 stays silent. The fixture carries no // want at all — this is the
// consent-webhook-trigger incident (2026-08-27, 28 diagnostics). A sqlx import
// in a _test.go file of that module does not turn the rule back on.
func TestNoSQLStack(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "grpcsvc/...")
}

// TestSQLStack — a module that really speaks SQL is judged as before: sqlsvc
// imports sqlx in its repository, appsql opens database/sql in its composition
// root (the verdict is per module, so an import outside the dal counts too).
func TestSQLStack(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "sqlsvc/...", "appsql/...")
}

// TestCustomSQLImports — settings.sql-imports names an in-house wrapper that
// hides the driver: the default stack does not see it, the setting turns the
// rule back on.
func TestCustomSQLImports(t *testing.T) {
	a := NewAnalyzer(Settings{SQLImports: []string{"gid.team/libs/pgstore"}})
	analysistest.Run(t, analysistest.TestData(), a, "wrapsvc/...")
}
