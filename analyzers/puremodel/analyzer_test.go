package puremodel

import (
	"go/token"
	"go/types"
	"testing"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "svc/...", "dalsvc/...")
}

func TestAnalyzerExclude(t *testing.T) {
	a := NewAnalyzer(Settings{
		Exclude: []string{"Legacy.Changed", "LegacyChanged"},
	})
	analysistest.Run(t, analysistest.TestData(), a, "exsvc/...")
}

// TestOwnModule — a method cannot be added to a type of a foreign module, so such
// a type is not an owner. The eval fixtures run in GOPATH mode without module
// information, hence the check is driven directly.
func TestOwnModule(t *testing.T) {
	named := func(path string) *types.Named {
		pkg := types.NewPackage(path, "model")
		obj := types.NewTypeName(token.NoPos, pkg, "Metric", nil)

		return types.NewNamed(obj, types.NewStruct(nil, nil), nil)
	}
	const root = "git.example.com/team/svc"
	tests := []struct {
		name   string
		module *analysis.Module
		path   string
		want   bool
	}{
		{"no module information", nil, "anything/domain/model", true},
		{"module without path", &analysis.Module{}, "anything/domain/model", true},
		{"module root", &analysis.Module{Path: root}, root, true},
		{"package of the module", &analysis.Module{Path: root}, root + "/internal/domain/model", true},
		{"foreign module", &analysis.Module{Path: root}, "git.example.com/team/lib/domain/model", false},
		{"sibling with the same prefix", &analysis.Module{Path: root}, root + "-lib/domain/model", false},
	}
	for i := range tests {
		tt := &tests[i]
		t.Run(tt.name, func(t *testing.T) {
			pass := &analysis.Pass{Module: tt.module}
			if got := ownModule(pass, named(tt.path)); got != tt.want {
				t.Fatalf("ownModule(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
