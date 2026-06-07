// Package nopanic реализует правило GID-161: panic используется только
// в пакете main (bootstrap). В остальном коде ошибки возвращаются
// и обрабатываются явно.
package nopanic

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-161"

// Analyzer — правило GID-161: panic используется только в пакете main.
var Analyzer = &analysis.Analyzer{
	Name: "gidnopanic",
	Doc:  ruleID + ": panic используется только в пакете main",
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
				return true // локальная функция panic — не встроенный panic
			}
			pass.Reportf(call.Pos(),
				"%s: panic используется только в пакете main — верните error", ruleID)
			return true
		})
	}
	return nil, nil
}
