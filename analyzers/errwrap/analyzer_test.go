package errwrap_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/errwrap"
)

func TestWrapAnalyzerBoundary(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), errwrap.WrapAnalyzer, "boundarysvc/...")
}

func TestWrapAnalyzerDomain(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), errwrap.WrapAnalyzer, "domainsvc/...")
}

func TestStaticAnalyzer(t *testing.T) {
	a := errwrap.NewStaticAnalyzer(errwrap.Settings{
		Exclude: []string{"gderror.NewUnhandledValueError"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "staticsvc/...")
}
