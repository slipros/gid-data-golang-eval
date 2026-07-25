package dirtree

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

// TestAnalyzer — the default canonical internal/ tree.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...")
}

// TestCustomTree — the tree from settings.tree replaces the default one
// and works at any level, not only in internal/.
func TestCustomTree(t *testing.T) {
	a := NewAnalyzer(Settings{
		Tree: map[string][]string{
			"pkg":     {"api", "contract"},
			"pkg/api": {"v1", "v2"},
		},
	})
	analysistest.Run(t, analysistest.TestData(), a, "custom/...")
}
