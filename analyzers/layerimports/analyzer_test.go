package layerimports_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/layerimports"
)

// TestAnalyzer runs the built-in matrix on testdata/src/svc/...:
//   - GID-132: dal -> domain, domain/model|usecase -> dal, service -> dal/repository;
//   - GID-170: domain|dal -> event;
//   - GID-172: client -> dal;
//   - GID-224: transport (server/schedule/validate/event) sees only domain/model;
//   - GID-225: app and transport leaves are imported by nobody;
//   - GID-226: metric is standalone, domain/dal do not import metric;
//   - GID-227: domain/model is the pure vocabulary;
//   - GID-228: domain/usecase does not import client;
//   - GID-229: client is isolated from the service layers.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), layerimports.Analyzer, "svc/...")
}

// TestAnalyzerSettings — settings.disable turns off a built-in rule,
// settings.rules adds a custom one (testdata/src/custom/...).
func TestAnalyzerSettings(t *testing.T) {
	a := layerimports.NewAnalyzer(layerimports.Settings{
		Disable: []string{"GID-224"},
		Rules: []layerimports.RuleSetting{{
			ID:     "SVC-1",
			Scope:  "domain/service",
			Banned: []string{"legacy"},
			Reason: "the legacy package is being removed",
		}},
	})
	analysistest.Run(t, analysistest.TestData(), a, "custom/...")
}
