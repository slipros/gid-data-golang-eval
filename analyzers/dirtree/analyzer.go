// Package dirtree implements rule GID-158: folder tree control.
// For each folder from the settings a list of allowed subfolders is defined;
// the appearance of a foreign folder is a warning (e.g. a new folder in
// internal/; perhaps it should be a service or usecase).
//
// The tree is configured in .golangci.yml (settings.tree); the key is a folder
// path (segments separated by /, matched anywhere in the import path), the
// value is the allowed subfolders. A tree given in settings replaces the default one.
package dirtree

import (
	"go/ast"
	"slices"
	"sort"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-158"

// defaultTree — the canonical service structure (ARCHITECTURE.md).
// Folders not listed as a key are not restricted. job holds background
// jobs (optional); schedule is a transport leaf (see GID-224/225).
var defaultTree = map[string][]string{
	"internal":                {"app", "client", "dal", "domain", "event", "job", "metric", "schedule", "server"},
	"internal/dal":            {"entity", "repository"},
	"internal/dal/repository": {"convert", "build"},
	"internal/domain":         {"model", "service", "usecase"},
	"internal/domain/service": {"convert"},
	"internal/server":         {"grpc", "http"},
}

// Analyzer — the variant with the default tree.
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Tree: "folder" -> allowed subfolders. Replaces the default tree.
	Tree map[string][]string `json:"tree"`
}

// NewAnalyzer builds the GID-158 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	tree := s.Tree
	if len(tree) == 0 {
		tree = defaultTree
	}
	return &analysis.Analyzer{
		Name: "giddirtree",
		Doc:  ruleID + ": a folder may contain only allowed subfolders (settings.tree). Fix: move the folder or add it to settings.tree",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, tree)
		},
	}
}

func run(pass *analysis.Pass, tree map[string][]string) (any, error) {
	pkgPath := pass.Pkg.Path()
	segs := pathseg.Segments(pkgPath)

	keys := make([]string, 0, len(tree))
	for key := range tree {
		keys = append(keys, key)
	}
	sort.Strings(keys) // a deterministic order of diagnostics

	for _, key := range keys {
		seq := pathseg.Segments(key)
		idx := pathseg.Index(pkgPath, seq...)
		if idx < 0 {
			continue
		}
		next := idx + len(seq)
		if next >= len(segs) {
			continue // the package is the key folder itself
		}
		if slices.Contains(tree[key], segs[next]) {
			continue
		}
		report(pass, key, segs[next], tree[key])
	}
	return nil, nil
}

func report(pass *analysis.Pass, key, dir string, allowed []string) {
	hint := ""
	if key == "internal" {
		hint = "; perhaps it should be a service or usecase"
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		pass.Reportf(file.Name.Pos(),
			"%s: folder %q is not allowed in %s/ (allowed: %s)%s; configure the tree via settings.tree",
			ruleID, dir, key, strings.Join(allowed, ", "), hint)
	}
}
