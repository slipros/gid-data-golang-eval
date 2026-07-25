package storageplace_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/storageplace"
)

// TestAnalyzer runs the default driver list and the default allowed layers on
// testdata/src/svc/...:
//   - positive: a key-value driver in /client, a SQL pool in /domain/service,
//     a driver in /job;
//   - negative: the driver in /dal/repository and in the composition root /app,
//     an external-API client (SMTP);
//   - boundary: /dal/entity (a pgtype column), a path that merely starts like a
//     driver prefix, a _test.go file outside dal;
//   - non-applicability: generated code.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), storageplace.Analyzer, "svc/...")
}

// TestAnalyzerSettings — the four settings on testdata/src/custom/...:
// settings.packages EXTENDS the default driver list (the in-house eredis
// wrapper is flagged in /client), settings.allow adds a layer (/job stays
// silent for a default driver), settings.exclude-packages drops an import from
// the driver list (a driver subpackage the built-in observability list does
// not know), settings.exclude-paths skips a package entirely.
func TestAnalyzerSettings(t *testing.T) {
	a := storageplace.NewAnalyzer(storageplace.Settings{
		Packages:        []string{"git.example.com/go-library/eredis"},
		Allow:           []string{"job"},
		ExcludePackages: []string{"github.com/redis/go-redis/pkg/instrumentation"},
		ExcludePaths:    []string{"legacy/cache"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "custom/...")
}
