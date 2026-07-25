package errwrap

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestWrapAnalyzerBoundary(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), WrapAnalyzer, "boundarysvc/...")
}

func TestWrapAnalyzerDomain(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), WrapAnalyzer, "domainsvc/...")
}

func TestWrapAnalyzerEvent(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), WrapAnalyzer, "eventsvc/...")
}

func TestStaticAnalyzer(t *testing.T) {
	a := NewStaticAnalyzer(Settings{
		Exclude: []string{"gderror.NewUnhandledValueError"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "staticsvc/...")
}

func TestRedundantStackAnalyzer(t *testing.T) {
	a := NewRedundantStackAnalyzer(Settings{
		Exclude: []string{"Service.excludedMethod"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "stacksvc/...")
}

func TestServiceMessageAnalyzer(t *testing.T) {
	a := NewServiceMessageAnalyzer(Settings{
		Exclude: []string{"Service.excludedMethod"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "servicemsgsvc/...")
}
