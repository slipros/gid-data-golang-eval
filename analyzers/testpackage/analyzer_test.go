package testpackage

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer — default settings:
//   - positive: blackbox/svc_test.go declares package blackbox_test;
//   - negative: svc/svc_test.go declares package svc and reaches the
//     unexported helper directly.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc", "blackbox")
}

// TestAnalyzerExclude — non-applicability: settings.exclude-paths exempts a
// directory that keeps a deliberate black-box suite.
func TestAnalyzerExclude(t *testing.T) {
	a := NewAnalyzer(Settings{
		ExcludePaths: []string{"exempt/blackbox"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "exempt/blackbox")
}
