// Package astwalk walks the package AST through the inspector shared by the
// whole golangci-lint run instead of a private ast.Inspect per rule.
//
// Why it matters. golangci-lint merges every go/analysis linter — ours and the
// upstream ones — into a single meta-linter with one action graph, and
// deduplicates actions by the pair (analyzer, package). So inspect.Analyzer,
// which many govet passes already require, is computed once per package and its
// inspector is handed to every rule that asks for it. A rule that instead calls
// ast.Inspect(file, …) pays for a full traversal of its own; with ~100 rules
// that is ~100 traversals of the same trees.
//
// The saving is in the filter. ast.Inspect visits every node and hands it to
// the callback; inspector.Preorder walks a flat, pre-built event list and skips
// whole subtrees that cannot contain a node of the requested type. Measured on
// a 174-file service: a full ast.Inspect pass is 0.40 ms, the same pass through
// the inspector filtered to *ast.CallExpr is 0.043 ms, filtered to
// *ast.FuncDecl — 0.004 ms.
//
// A rule opts in by adding Requires: astwalk.Requires to its analysis.Analyzer
// and calling NodesOf (one node type) or Nodes/Around/NodesPruning (several)
// instead of ast.Inspect.
package astwalk

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Requires is the Requires list an analyzer needs in order to use this package.
var Requires = []*analysis.Analyzer{inspect.Analyzer}

// fileNode is prepended to every filter so that the walkers can tell which file
// the nodes that follow belong to: the inspector reports nodes in source order,
// so an *ast.File event opens the run of nodes of that file.
var fileNode = (*ast.File)(nil)

// Inspector returns the shared inspector of the package under analysis, or nil
// when the analyzer forgot to declare Requires: astwalk.Requires. That is a
// wiring bug in the rule; the walkers below then visit nothing, so the rule
// stays silent instead of taking the whole golangci-lint run down with it.
func Inspector(pass *analysis.Pass) *inspector.Inspector {
	insp, ok := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	if !ok {
		return nil
	}

	return insp
}

// NodesOf calls visit for every node of type T in the package, passing the file
// the node came from. It is the form to reach for when a rule judges one kind
// of node: the node arrives already typed, so no rule repeats the type
// assertion the filter has just guaranteed.
//
// skip, when not nil, is asked once per file: a file it rejects contributes no
// nodes at all. That replaces the usual `for _, file := range pass.Files { if
// isGenerated(file) { continue }; ast.Inspect(file, …) }` preamble and keeps
// the per-file decision out of the hot callback.
func NodesOf[T ast.Node](pass *analysis.Pass, skip func(*ast.File) bool, visit func(*ast.File, T)) {
	var want T

	Nodes(pass, []ast.Node{want}, skip, func(file *ast.File, n ast.Node) {
		if node, ok := n.(T); ok {
			visit(file, node)
		}
	})
}

// Nodes is NodesOf for a rule that judges several kinds of node at once and
// tells them apart with a type switch.
func Nodes(pass *analysis.Pass, filter []ast.Node, skip func(*ast.File) bool, visit func(*ast.File, ast.Node)) {
	walk(pass, filter, skip, func(file *ast.File, n ast.Node, push bool) bool {
		if push {
			visit(file, n)
		}

		return true
	})
}

// Around is Nodes with both entry and exit events, for a rule that needs to
// know it is *inside* something — a loop body, a function body — rather than
// just that the node exists. Tracking that with a counter incremented on entry
// and decremented on exit replaces the usual "collect all the enclosing nodes
// first, then test every candidate against the list", which is quadratic.
//
// visit is called with push=true on the way in and push=false on the way out;
// its return value is honoured on entry only, and false skips the subtree.
func Around(pass *analysis.Pass, filter []ast.Node, skip func(*ast.File) bool, visit func(*ast.File, ast.Node, bool) bool) {
	walk(pass, filter, skip, visit)
}

// NodesPruning is Nodes for a rule that reports the outermost match and must
// not descend into it — an entity literal filled inside another entity literal
// is one violation, not two. visit returns whether to walk the subtree.
func NodesPruning(pass *analysis.Pass, filter []ast.Node, skip func(*ast.File) bool, visit func(*ast.File, ast.Node) bool) {
	walk(pass, filter, skip, func(file *ast.File, n ast.Node, push bool) bool {
		if !push {
			return true
		}

		return visit(file, n)
	})
}

// walk drives the shared inspector for all the forms above: it tracks the
// current file, applies skip once per file, and hands everything else to visit.
func walk(pass *analysis.Pass, filter []ast.Node, skip func(*ast.File) bool, visit func(*ast.File, ast.Node, bool) bool) {
	insp := Inspector(pass)
	if insp == nil {
		return
	}

	withFile := make([]ast.Node, 0, len(filter)+1)
	withFile = append(withFile, fileNode)
	withFile = append(withFile, filter...)

	var (
		current *ast.File
		skipped bool
	)

	insp.Nodes(withFile, func(n ast.Node, push bool) bool {
		if file, ok := n.(*ast.File); ok {
			if push {
				current = file
				skipped = skip != nil && skip(file)
			}

			return !skipped
		}
		if skipped {
			return true
		}

		return visit(current, n, push)
	})
}
