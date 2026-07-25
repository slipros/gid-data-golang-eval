package utilpkg

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer — the default blacklist: util/helpers/common are caught,
// convert/stringutil/model are not.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer,
		"util", "helpers", "common", "convert", "stringutil", "model")
}

// TestCustomNames — settings.names replaces the default: junk is matched,
// while util no longer is.
func TestCustomNames(t *testing.T) {
	a := NewAnalyzer(Settings{Names: []string{"junk"}})
	analysistest.Run(t, analysistest.TestData(), a, "junk", "customutil")
}
