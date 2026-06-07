package enumstring_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/enumstring"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), enumstring.Analyzer, "enumstring")
}

func TestBasedAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), enumstring.BasedAnalyzer, "enumbased/...")
}
