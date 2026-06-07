// Package onlypkgerrors реализует правило GID-146: для работы с ошибками
// используется только github.com/pkg/errors. Создание ошибок через
// стандартные errors.New и fmt.Errorf запрещено везде.
//
// Проверка цепочки ошибок — std errors.Is/As/Unwrap — не создание,
// она разрешена (у pkg/errors этих функций нет).
package onlypkgerrors

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"
)

const (
	ruleID     = "GID-146"
	allowedPkg = "github.com/pkg/errors"
)

// forbidden — std-конструкторы ошибок: пакет -> функции.
var forbidden = map[string]map[string]struct{}{
	"errors": {"New": {}, "Join": {}},
	"fmt":    {"Errorf": {}},
}

// Analyzer — правило GID: errors are created only via .
var Analyzer = &analysis.Analyzer{
	Name: "gidonlypkgerrors",
	Doc:  ruleID + ": errors are created only via " + allowedPkg,
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
			f, ok := typeutil.Callee(pass.TypesInfo, call).(*types.Func)
			if !ok || f.Pkg() == nil {
				return true
			}
			fPkg := f.Pkg()
			names, ok := forbidden[fPkg.Path()]
			if !ok {
				return true
			}
			if _, ok := names[f.Name()]; !ok {
				return true
			}
			pass.Reportf(call.Pos(),
				"%s: %s.%s is forbidden. Fix: use only %s for errors",
				ruleID, fPkg.Name(), f.Name(), allowedPkg)
			return true
		})
	}
	return nil, nil
}
