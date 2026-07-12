package errmapfunc_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/errmapfunc"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), errmapfunc.Analyzer, "svc")
}

// TestCustomPackages — a project-configured errors facade (settings.packages)
// is recognized as a classifier package: the facade mapper is flagged, the
// facade bool-predicate stays clean. Under the default whitelist the same
// package produces no diagnostics (myerrors is neither "errors" nor
// github.com/pkg/errors) — proving the setting, not a hardcoded list, drives it.
func TestCustomPackages(t *testing.T) {
	a := errmapfunc.NewAnalyzer(errmapfunc.Settings{Packages: []string{"myerrors"}})
	analysistest.Run(t, analysistest.TestData(), a, "customfacade")
}
