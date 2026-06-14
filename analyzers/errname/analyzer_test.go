package errname_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/errname"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), errname.Analyzer, "svc/...")
}

func TestAnalyzerCustomSettings(t *testing.T) {
	a := errname.NewAnalyzer(errname.Settings{
		Names:   []string{"ErrOops", "ErrLegacy"},
		Exclude: []string{"ErrLegacy"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "custom/...")
}
