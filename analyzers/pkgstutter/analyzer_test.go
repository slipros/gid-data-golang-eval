package pkgstutter_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/pkgstutter"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), pkgstutter.Analyzer, "widget", "log", "mainpkg", "repository", "service", "story")
}
