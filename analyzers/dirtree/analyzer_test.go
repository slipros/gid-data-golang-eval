package dirtree_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/slipros/gid-data-golang-eval/analyzers/dirtree"
)

// TestAnalyzer — дефолтное каноничное дерево internal/.
func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), dirtree.Analyzer, "svc/...")
}

// TestCustomTree — дерево из settings.tree заменяет дефолтное
// и работает на любом уровне, не только в internal/.
func TestCustomTree(t *testing.T) {
	a := dirtree.NewAnalyzer(dirtree.Settings{
		Tree: map[string][]string{
			"pkg":     {"api", "contract"},
			"pkg/api": {"v1", "v2"},
		},
	})
	analysistest.Run(t, analysistest.TestData(), a, "custom/...")
}
