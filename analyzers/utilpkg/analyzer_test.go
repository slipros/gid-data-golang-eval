package utilpkg_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/utilpkg"
)

// TestAnalyzer — дефолтный чёрный список: util/helpers/common ловятся,
// convert/stringutil/model — нет.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), utilpkg.Analyzer,
		"util", "helpers", "common", "convert", "stringutil", "model")
}

// TestCustomNames — settings.names замещает дефолт: junk матчится,
// а util — уже нет.
func TestCustomNames(t *testing.T) {
	a := utilpkg.NewAnalyzer(utilpkg.Settings{Names: []string{"junk"}})
	analysistest.Run(t, analysistest.TestData(), a, "junk", "customutil")
}
