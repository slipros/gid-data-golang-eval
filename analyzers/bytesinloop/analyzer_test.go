package bytesinloop_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/bytesinloop"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), bytesinloop.Analyzer, "bytesinloop", "noloop")
}
