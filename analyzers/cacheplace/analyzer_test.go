package cacheplace_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/cacheplace"
)

// TestAnalyzer — дефолтный список кэш-библиотек.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), cacheplace.Analyzer, "svc/...")
}

// TestCustomPackages — settings.packages заменяет дефолтный список.
func TestCustomPackages(t *testing.T) {
	a := cacheplace.NewAnalyzer(cacheplace.Settings{
		Packages: []string{"example.com/inhouse/cache"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "custom/...")
}
