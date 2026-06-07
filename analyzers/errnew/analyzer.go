// Package errnew реализует правило GID-136 (линтер giderrnew):
// errors.New из github.com/pkg/errors допустим только в объявлении
// package-level var — статичные ошибки объявляются заранее (ErrX), а
// не конструируются в рантайме.
//
// Вызов errors.New внутри тела функции, метода или func-литерала —
// диагностика. Объявление package-level var (включая var-блоки)
// ErrX = errors.New("...") — норма.
//
// Вне зоны правила:
//   - errors.Errorf — динамический контекст легитимен, его место
//     регулируют GID-144/GID-145;
//   - стандартный errors.New — он уже запрещён GID-146;
//   - errors.New из любого другого (не github.com/pkg/errors) пакета.
//
// pkg/errors определяется по import-пути github.com/pkg/errors через
// TypesInfo. Сгенерированный код (ast.IsGenerated) пропускается.
package errnew

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"
)

const ruleID = "GID-136"

// Analyzer — правило GID-136: errors.New (pkg/errors) только в package-level var.
var Analyzer = &analysis.Analyzer{
	Name: "giderrnew",
	Doc:  ruleID + ": errors.New (pkg/errors) только в package-level var, не в рантайме",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		checkFile(pass, file)
	}
	return nil, nil
}

// checkFile обходит тела всех функций, методов и func-литералов файла и
// репортит в них вызовы errors.New из pkg/errors. Вызовы вне тел функций
// (package-level var ErrX = errors.New(...)) не задеваются.
//
// Тело func-литерала — рантайм даже когда сам литерал записан в
// package-level var: errors.New там вычисляется при вызове литерала.
func checkFile(pass *analysis.Pass, file *ast.File) {
	var bodies []*ast.BlockStmt

	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			if node.Body != nil {
				bodies = append(bodies, node.Body)
			}
		case *ast.FuncLit:
			bodies = append(bodies, node.Body)
		}
		return true
	})

	for _, body := range bodies {
		ast.Inspect(body, func(n ast.Node) bool {
			// Не спускаемся во вложенный func-литерал — его тело обходится
			// отдельной итерацией, иначе вызов отрепортится дважды.
			if _, ok := n.(*ast.FuncLit); ok {
				return false
			}
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if isPkgErrorsNew(pass, call) {
				pass.Reportf(call.Pos(),
					"%s: errors.New в рантайме — объявите package-level var ErrX (см. GID-169: error.go)",
					ruleID)
			}
			return true
		})
	}
}

// isPkgErrorsNew сообщает, является ли call вызовом errors.New из
// github.com/pkg/errors.
func isPkgErrorsNew(pass *analysis.Pass, call *ast.CallExpr) bool {
	const pkgErrorsPath = "github.com/pkg/errors"
	fn := typeutil.Callee(pass.TypesInfo, call)
	f, ok := fn.(*types.Func)
	if !ok || f.Pkg() == nil {
		return false
	}
	pkg := f.Pkg()
	return pkg.Path() == pkgErrorsPath && f.Name() == "New"
}
