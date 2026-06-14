package exitonce_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/exitonce"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), exitonce.Analyzer,
		"mainhelper",  // positive (os.Exit in a helper of the main package) + negative (one os.Exit in main, returning an error)
		"libpkg",      // positive (log.Fatal/logrus.Fatal* in a non-main package) + negative (returning an error)
		"twoexit",     // positive (two os.Exit in main — a duplicate)
		"okmain",      // boundary (defer + one os.Exit — fine)
		"closuremain", // boundary (os.Exit in a closure inside main — counts as a call in main)
		"cleanlib",    // inapplicability (a library package without exit calls)
	)
}
