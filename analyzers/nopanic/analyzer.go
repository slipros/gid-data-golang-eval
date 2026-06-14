// Package nopanic implements rule GID-161: panic is used only in package
// main (bootstrap). In all other code errors are returned and handled
// explicitly.
package nopanic

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-161"

// Analyzer — rule GID-161: panic is used only in package main. Fix: return an error instead.
var Analyzer = &analysis.Analyzer{
	Name: "gidnopanic",
	Doc:  ruleID + ": panic is used only in package main. Fix: return an error instead",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	if pass.Pkg.Name() == "main" {
		return nil, nil
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			ident, ok := call.Fun.(*ast.Ident)
			if !ok || ident.Name != "panic" {
				return true
			}
			if _, isBuiltin := pass.TypesInfo.Uses[ident].(*types.Builtin); !isBuiltin {
				return true // a local function named panic is not the builtin panic
			}
			pass.Reportf(call.Pos(),
				"%s: panic is allowed only in package main. Fix: return an error instead", ruleID)
			return true
		})
	}
	return nil, nil
}
