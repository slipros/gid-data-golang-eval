// Package ctornaming реализует правило GID-104: конструктор именуется
// New<Entity> (NewHello, NewPlaceOrder). Голый New не подходит — все
// сущности слоя живут в одном пакете, будет конфликт имён.
//
// Исключение: composition root (internal/app/...) — там по шаблону
// живёт функция New() приложения.
package ctornaming

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-104"

// Analyzer — правило GID-104: a constructor must be named New<Entity>, not bare New. Fix: rename New to New<Entity>.
var Analyzer = &analysis.Analyzer{
	Name: "gidctor",
	Doc:  ruleID + ": a constructor must be named New<Entity>, not bare New. Fix: rename New to New<Entity>",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	if pathseg.Contains(pass.Pkg.Path(), "app") {
		return nil, nil
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Name.Name != "New" {
				continue
			}
			pass.Reportf(fn.Name.Pos(),
				"%s: a constructor must be named New<Entity>, not bare New. Fix: rename it to New<Entity> (bare New clashes with other entities in the package)",
				ruleID)
		}
	}
	return nil, nil
}
