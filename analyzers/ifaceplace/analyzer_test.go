package ifaceplace

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...")
}

// TestExclude — a port file excluded via settings.exclude (by file name or
// by interface name) is not reported.
func TestExclude(t *testing.T) {
	t.Run("by_file_name", func(t *testing.T) {
		a := NewAnalyzer(Settings{Exclude: []string{"ports.go"}})
		analysistest.Run(t, analysistest.TestData(), a, "portfile")
	})
	t.Run("by_interface_name", func(t *testing.T) {
		a := NewAnalyzer(Settings{Exclude: []string{"Loader"}})
		analysistest.Run(t, analysistest.TestData(), a, "portfile")
	})
}
