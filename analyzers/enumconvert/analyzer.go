// Package enumconvert реализует правило GID-143 (линтер gidenumconvert):
// map-конвертация enum обязана обрабатывать отсутствующий ключ через
// gderror.NewUnhandledValueError.
//
// Действует только в convert-пакетах (последний сегмент пути — convert).
// Детектируется индексация мапы m[key], у которой тип ключа — именованный
// тип с underlying string (enum по GID-123), а тип значения — тоже
// именованный тип (конвертация enum→enum / enum→модельный тип):
//
//   - индексация не в comma-ok форме (одиночное присваивание / выражение) —
//     отсутствующий ключ молча даёт zero-value, его нельзя обработать;
//   - comma-ok форма есть, но в теле той же функции нет вызова
//     gderror.NewUnhandledValueError — отсутствующий ключ не обрабатывается.
//
// Мапы с базовыми ключами (string, int) не матчятся. Вне convert-пакетов
// не матчится. Сгенерированный код (ast.IsGenerated) пропускается.
// LoadMode — TypesInfo (нужны типы ключа/значения).
package enumconvert

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-143"

// Analyzer — правило GID-143: enum-конвертация через map обрабатывает
// отсутствующий ключ через gderror.NewUnhandledValueError.
var Analyzer = &analysis.Analyzer{
	Name: "gidenumconvert",
	Doc:  ruleID + ": map-конвертация enum обрабатывает отсутствующий ключ через gderror.NewUnhandledValueError",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	// Scope: только convert-пакеты (последний сегмент пути).
	if !pathseg.EndsWith(pass.Pkg.Path(), "convert") {
		return nil, nil
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			checkFunc(pass, fn)
		}
	}
	return nil, nil
}

// checkFunc проверяет все enum-индексации мапы в теле функции.
func checkFunc(pass *analysis.Pass, fn *ast.FuncDecl) {
	hasHandler := callsUnhandledValueError(pass, fn.Body)
	commaOk := commaOkIndexes(fn.Body)
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		idx, ok := n.(*ast.IndexExpr)
		if !ok {
			return true
		}
		if !isEnumMapIndex(pass, idx) {
			return true
		}
		if _, ok := commaOk[idx]; ok {
			// comma-ok форма есть — нужен явный вызов обработчика в этой же функции.
			if !hasHandler {
				pass.Reportf(idx.Pos(),
					"%s: отсутствующий ключ enum-конвертации обрабатывается gderror.NewUnhandledValueError",
					ruleID)
			}
			return true
		}
		// Не comma-ok: отсутствующий ключ молча даёт zero-value.
		pass.Reportf(idx.Pos(),
			"%s: enum-конвертация через map без comma-ok — "+
				"отсутствующий ключ должен давать gderror.NewUnhandledValueError",
			ruleID)
		return true
	})
}

// isEnumMapIndex сообщает, что idx — индексация мапы, у которой ключ —
// именованный string-тип (enum), а значение — именованный тип.
func isEnumMapIndex(pass *analysis.Pass, idx *ast.IndexExpr) bool {
	t := pass.TypesInfo.TypeOf(idx.X)
	if t == nil {
		return false
	}
	m, ok := t.Underlying().(*types.Map)
	if !ok {
		return false
	}
	return isNamedString(m.Key()) && isNamed(m.Elem())
}

// isNamedString: именованный тип с underlying string (enum по GID-123).
func isNamedString(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	basic, ok := named.Underlying().(*types.Basic)
	return ok && basic.Kind() == types.String
}

// isNamed: именованный тип (enum→enum / enum→модельный тип).
func isNamed(t types.Type) bool {
	_, ok := t.(*types.Named)
	return ok
}

// commaOkIndexes собирает индексации мапы, использованные в comma-ok форме
// (v, ok := m[k] / v, ok = m[k]) — RHS из одного выражения при двух LHS.
func commaOkIndexes(body *ast.BlockStmt) map[*ast.IndexExpr]struct{} {
	out := map[*ast.IndexExpr]struct{}{}
	ast.Inspect(body, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok || len(assign.Lhs) != 2 || len(assign.Rhs) != 1 {
			return true
		}
		if idx, ok := assign.Rhs[0].(*ast.IndexExpr); ok {
			out[idx] = struct{}{}
		}
		return true
	})
	return out
}

// callsUnhandledValueError сообщает, что в теле есть вызов
// gderror.NewUnhandledValueError.
func callsUnhandledValueError(pass *analysis.Pass, body *ast.BlockStmt) bool {
	const (
		// gderrorPath — import-путь внутренней библиотеки ошибок.
		gderrorPath = "gitlab.gid.team/gid-data/tech/golang/libs/helper.git/errors"
		// unhandledCtor — конструктор обработки отсутствующего ключа.
		unhandledCtor = "NewUnhandledValueError"
	)
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		fn, ok := typeutil.Callee(pass.TypesInfo, call).(*types.Func)
		if !ok || fn.Pkg() == nil {
			return true
		}
		fnPkg := fn.Pkg()
		if fnPkg.Path() == gderrorPath && fn.Name() == unhandledCtor {
			found = true
			return false
		}
		return true
	})
	return found
}
