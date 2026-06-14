// Package chanbuf implements rule GID-179 (Uber: channel size is one or none):
// the channel buffer size in make(chan T, N) may only be 0 or 1.
// A larger buffer (N > 1) with a constant value is forbidden — it almost always
// masks a synchronization problem and must be justified explicitly.
//
// What is matched:
//   - make(chan T, 2), make(chan T, 100) — a literal > 1;
//   - make(chan T, maxWorkers), where maxWorkers is a named const = 10
//     (the value is computed via TypesInfo, constant.Int).
//
// What is NOT matched:
//   - make(chan T), make(chan T, 0), make(chan T, 1) — buffer 0 or 1;
//   - make(chan T, n), where n is a variable/call (the size is justified at
//     runtime — left to review);
//   - make([]T, N), make(map[K]V, N) — not channels.
//
// Targeted opt-out: //nolint:gidchanbuf (works via golangci-lint,
// nothing is required in the analyzer code).
package chanbuf

import (
	"go/ast"
	"go/constant"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-179"

// Analyzer — rule GID-179: channel buffer size may only be 0 or 1.
var Analyzer = &analysis.Analyzer{
	Name: "gidchanbuf",
	Doc:  ruleID + ": channel buffer size must be 0 or 1. Fix: use an unbuffered channel or buffer 1, or justify a larger buffer with //nolint:gidchanbuf.",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if !isMakeBuiltin(pass, call) {
				return true
			}
			// make(chan T, N): the first argument is the channel type, the second is the size.
			if len(call.Args) < 2 {
				return true
			}
			if _, ok := call.Args[0].(*ast.ChanType); !ok {
				return true // make([]T, N) / make(map[K]V, N) — not a channel.
			}
			sizeExpr := call.Args[1]
			tv, ok := pass.TypesInfo.Types[sizeExpr]
			if !ok || tv.Value == nil {
				return true // the size is not a constant (a variable/call) — skip.
			}
			size, ok := constant.Int64Val(constant.ToInt(tv.Value))
			if !ok {
				return true
			}
			if size <= 1 {
				return true // 0 and 1 are allowed.
			}
			pass.Reportf(sizeExpr.Pos(),
				"%s: channel buffer %d is not allowed (only 0 or 1). "+
					"Fix: use an unbuffered channel or buffer 1, or justify a larger buffer with //nolint:gidchanbuf.",
				ruleID, size)
			return true
		})
	}
	return nil, nil
}

// isMakeBuiltin: the call is the built-in make, not a local function named make.
func isMakeBuiltin(pass *analysis.Pass, call *ast.CallExpr) bool {
	ident, ok := call.Fun.(*ast.Ident)
	if !ok || ident.Name != "make" {
		return false
	}
	builtin, ok := pass.TypesInfo.Uses[ident].(*types.Builtin)
	return ok && builtin.Name() == "make"
}
